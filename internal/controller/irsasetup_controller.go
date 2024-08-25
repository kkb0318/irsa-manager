/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	irsav1alpha1 "github.com/kkb0318/irsa-manager/api/v1alpha1"
	awsclient "github.com/kkb0318/irsa-manager/internal/aws"
	"github.com/kkb0318/irsa-manager/internal/eks"
	"github.com/kkb0318/irsa-manager/internal/handler"
	"github.com/kkb0318/irsa-manager/internal/issuer"
	"github.com/kkb0318/irsa-manager/internal/kubernetes"
	"github.com/kkb0318/irsa-manager/internal/manifests"
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
	"github.com/kkb0318/irsa-manager/internal/selfhosted/oidc"
	"github.com/kkb0318/irsa-manager/internal/selfhosted/webhook"
)

const irsamanagerFinalizer = "irsa-manager.kkb0318.github.io/finalizers"

// IRSASetupReconciler reconciles a IRSASetup object
type IRSASetupReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	AwsClient awsclient.AwsClient
}

//+kubebuilder:rbac:groups=irsa-manager.kkb0318.github.io,resources=irsasetups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=irsa-manager.kkb0318.github.io,resources=irsasetups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=irsa-manager.kkb0318.github.io,resources=irsasetups/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="certificates.k8s.io",resources=certificatesigningrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *IRSASetupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	obj := &irsav1alpha1.IRSASetup{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if r.AwsClient == nil {
		awsClient, err := awsclient.NewAwsClientFactory(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
		r.AwsClient = awsClient
	}
	kubeClient, err := kubernetes.NewKubernetesClient(r.Client, kubernetes.Owner{Field: "irsa-manager"})
	if err != nil {
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(obj, irsamanagerFinalizer) {
		controllerutil.AddFinalizer(obj, irsamanagerFinalizer)
		if err := r.Update(ctx, obj); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	defer func() {
		if e := r.Get(ctx, req.NamespacedName, &irsav1alpha1.IRSASetup{}); e != nil {
			return
		}
		statusHandler := handler.NewStatusHandler(kubeClient)
		if e := statusHandler.Patch(ctx, obj); e != nil {
			return
		}
	}()

	if !obj.DeletionTimestamp.IsZero() {
		if obj.Spec.Mode == irsav1alpha1.ModeEks {
			err = r.reconcileDeleteEks()
		} else {
			err = r.reconcileDeleteSelfhosted(ctx, obj, kubeClient)
		}
		if err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(obj, irsamanagerFinalizer)
		err = r.Update(ctx, obj)
		if err == nil {
			log.Info("successfully deleted")
		}
		return ctrl.Result{}, err
	}

	if err := r.reconcile(ctx, obj, kubeClient); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("successfully reconciled")
	return ctrl.Result{}, nil
}

func (r *IRSASetupReconciler) reconcile(ctx context.Context, obj *irsav1alpha1.IRSASetup, kubeClient *kubernetes.KubernetesClient) error {
	if obj.Spec.Mode == irsav1alpha1.ModeEks {
		return reconcileEks(ctx, obj)
	}
	return reconcileSelfhosted(ctx, obj, r.AwsClient, kubeClient)
}

func (r *IRSASetupReconciler) reconcileDeleteEks() error {
	return nil
}

func (r *IRSASetupReconciler) reconcileDeleteSelfhosted(ctx context.Context, obj *irsav1alpha1.IRSASetup, kubeClient *kubernetes.KubernetesClient) error {
	if !obj.Spec.Cleanup {
		return nil
	}
	factory, err := newOIDCIdpFactory(ctx, obj, nil, r.AwsClient)
	if err != nil {
		return err
	}
	secret, err := manifests.NewSecretBuilder().Build(manifests.SshKeyNamespacedName())
	if err != nil {
		return err
	}
	kubeHandler := handler.NewKubernetesHandler(kubeClient)
	kubeHandler.Append(secret)
	webhookSetup, err := webhook.NewWebHookSetup()
	if err != nil {
		return err
	}
	for _, r := range webhookSetup.Resources() {
		kubeHandler.Append(r)
	}
	_, err = kubeHandler.DeleteAll(ctx)
	if err != nil {
		return err
	}
	issuerMeta, err := issuer.NewOIDCIssuerMeta(obj)
	if err != nil {
		return err
	}
	return selfhosted.Delete(
		ctx,
		factory,
		issuerMeta,
	)
}

// reconcileSelfhosted ensures that the self-hosted resources are set up correctly.
// This function performs the following operations based on the state of the object:
// - If the self-hosted setup has previously succeeded, the function returns immediately without making changes.
// - If the self-hosted setup was previously attempted but failed, or if it's being run for the first time, it will attempt to create all necessary resources. This includes the creation of key pairs, JWKs, OIDC IDP configurations, and Kubernetes secrets.
// - The function enforces a 'force update' strategy in case of failures related to kubernetes Secrets creation or OIDC setup. This means it starts from scratch to ensure all components are correctly configured.
func reconcileSelfhosted(ctx context.Context, obj *irsav1alpha1.IRSASetup, awsClient awsclient.AwsClient, kubeClient *kubernetes.KubernetesClient) error {
	log := ctrllog.FromContext(ctx)
	if irsav1alpha1.IsReadyConditionTrue(*obj) {
		// Selfhosted Setup have already succeeded
		log.Info("the self-hosted resources have already set up")
		return nil
	}
	log.Info("the self-hosted resources are setting up")
	keyPair, err := selfhosted.CreateKeyPair()
	if err != nil {
		return err
	}
	jwk, err := selfhosted.NewJWK(keyPair.PublicKey())
	if err != nil {
		return err
	}
	factory, err := newOIDCIdpFactory(ctx, obj, jwk, awsClient)
	if err != nil {
		return err
	}
	secret, err := manifests.NewSecretBuilder().WithSSHKey(*keyPair).Build(manifests.SshKeyNamespacedName())
	if err != nil {
		return err
	}
	kubeHandlerForOidc := handler.NewKubernetesHandler(kubeClient)
	kubeHandlerForOidc.Append(secret)

	// for webhook setup
	webhookSetup, err := webhook.NewWebHookSetup()
	if err != nil {
		return err
	}

	// e is set only when an error occurs in an external dependency process and is reflected in the CRs status
	var e error
	var reason irsav1alpha1.SelfhostedConditionReason
	defer func() {
		if e != nil {
			*obj = irsav1alpha1.StatusNotReady(*obj, string(reason), e.Error())
		}
	}()

	forceUpdate := irsav1alpha1.HasConditionReason(
		irsav1alpha1.ReadyStatus(*obj),
		string(irsav1alpha1.SelfHostedReasonFailedKeys),
		string(irsav1alpha1.SelfHostedReasonFailedOidc),
	)
	issuerMeta, err := issuer.NewOIDCIssuerMeta(obj)
	if err != nil {
		e = err
		reason = irsav1alpha1.SelfHostedReasonFailedIssuer
		return err
	}
	err = selfhosted.Execute(
		ctx,
		factory,
		issuerMeta,
		forceUpdate,
	)
	if err != nil {
		e = err
		reason = irsav1alpha1.SelfHostedReasonFailedOidc
		return err
	}
	if forceUpdate {
		_, err = kubeHandlerForOidc.ApplyAll(ctx)
	} else {
		err = kubeHandlerForOidc.CreateAll(ctx)
	}
	if err != nil {
		e = err
		reason = irsav1alpha1.SelfHostedReasonFailedKeys
		return err
	}
	// for webhook update
	kubeHandlerForWebhook := handler.NewKubernetesHandler(kubeClient)
	for _, r := range webhookSetup.Resources() {
		kubeHandlerForWebhook.Append(r)
	}
	_, err = kubeHandlerForWebhook.ApplyAll(ctx)
	if err != nil {
		e = err
		reason = irsav1alpha1.SelfHostedReasonFailedWebhook
		return err
	}
	*obj = irsav1alpha1.SetupStatusReady(*obj, string(irsav1alpha1.SelfHostedReasonReady), "successfully setup resources for self-hosted")
	log.Info("the self-hosted resources have successfully set up")
	return nil
}

// reconcileEks iterates tasks for EKS mode.
func reconcileEks(ctx context.Context, obj *irsav1alpha1.IRSASetup) error {
	log := ctrllog.FromContext(ctx)
	var reason irsav1alpha1.EksConditionReason

	// e is set only when an error occurs in an external dependency process and is reflected in the CRs status
	var e error
	defer func() {
		if e != nil {
			*obj = irsav1alpha1.StatusNotReady(*obj, string(reason), e.Error())
		}
	}()
	err := eks.Validate(obj)
	if err != nil {
		e = err
		reason = irsav1alpha1.EksNotReady
		return err
	}
	*obj = irsav1alpha1.SetupStatusReady(*obj, string(irsav1alpha1.EksReasonReady), "successfully setup for eks")
	log.Info("The OIDC for EKS has been successfully set up")
	return nil
}

func newOIDCIdpFactory(ctx context.Context, obj *irsav1alpha1.IRSASetup, jwk *selfhosted.JWK, awsClient awsclient.AwsClient) (selfhosted.OIDCIdPFactory, error) {
	region := obj.Spec.Discovery.S3.Region
	bucketName := obj.Spec.Discovery.S3.BucketName
	jwksFileName := "keys.json"
	factory, err := oidc.NewAwsS3IdpFactory(
		ctx,
		region,
		bucketName,
		jwk,
		jwksFileName,
		awsClient,
	)
	if err != nil {
		return nil, err
	}
	return factory, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IRSASetupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&irsav1alpha1.IRSASetup{}).
		Complete(r)
}
