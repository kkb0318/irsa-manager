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
	awsclient "github.com/kkb0318/irsa-manager/internal/client"
	"github.com/kkb0318/irsa-manager/internal/handler"
	"github.com/kkb0318/irsa-manager/internal/kubernetes"
	"github.com/kkb0318/irsa-manager/internal/manifests"
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
	"github.com/kkb0318/irsa-manager/internal/selfhosted/oidc"
)

const irsamanagerFinalizer = "irsa.kkb0318.github.io/finalizers"

// IRSASetupReconciler reconciles a IRSASetup object
type IRSASetupReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	AwsClient awsclient.AwsClient
}

//+kubebuilder:rbac:groups=irsa.kkb0318.github.io,resources=irsasetups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=irsa.kkb0318.github.io,resources=irsasetups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=irsa.kkb0318.github.io,resources=irsasetups/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IRSASetup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
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

	if !obj.DeletionTimestamp.IsZero() {
		err = r.reconcileDelete(ctx, obj, kubeClient)
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
	err := reconcileSelfhosted(ctx, obj, r.AwsClient, kubeClient)
	return err
}

func (r *IRSASetupReconciler) reconcileDelete(ctx context.Context, obj *irsav1alpha1.IRSASetup, kubeClient *kubernetes.KubernetesClient) error {
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
	err = kubeHandler.DeleteAll(ctx)
	if err != nil {
		return err
	}
	err = selfhosted.Delete(ctx, factory)
	if err != nil {
		return err
	}
	controllerutil.RemoveFinalizer(obj, irsamanagerFinalizer)
	return r.Update(ctx, obj)
}

// reconcileSelfhosted ensures that the self-hosted resources are set up correctly.
// This function performs the following operations based on the state of the object:
// - If the self-hosted setup has previously succeeded, the function returns immediately without making changes.
// - If the self-hosted setup was previously attempted but failed, or if it's being run for the first time, it will attempt to create all necessary resources. This includes the creation of key pairs, JWKs, OIDC IDP configurations, and Kubernetes secrets.
// - The function enforces a 'force update' strategy in case of failures related to kubernetes Secrets creation or OIDC setup. This means it starts from scratch to ensure all components are correctly configured.
func reconcileSelfhosted(ctx context.Context, obj *irsav1alpha1.IRSASetup, awsClient awsclient.AwsClient, kubeClient *kubernetes.KubernetesClient) error {
	log := ctrllog.FromContext(ctx)
	if irsav1alpha1.IsSelfHostedReadyConditionTrue(*obj) {
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
	kubeHandler := handler.NewKubernetesHandler(kubeClient)
	kubeHandler.Append(secret)

	var e error
	var reason irsav1alpha1.SelfHostedReason
	defer func() {
		if e != nil {
			*obj = irsav1alpha1.SelfHostedStatusNotReady(*obj, string(reason), e.Error())
		}
	}()

	forceUpdate := irsav1alpha1.HasConditionReason(
		irsav1alpha1.SelfHostedReadyStatus(*obj),
		string(irsav1alpha1.SelfHostedReasonFailedKeys),
		string(irsav1alpha1.SelfHostedReasonFailedOidc),
	)
	err = selfhosted.Execute(ctx, factory, forceUpdate)
	if err != nil {
		e = err
		reason = irsav1alpha1.SelfHostedReasonFailedOidc
		return err
	}
	if forceUpdate {
		err = kubeHandler.ApplyAll(ctx)
	} else {
		err = kubeHandler.CreateAll(ctx)
	}
	if err != nil {
		e = err
		reason = irsav1alpha1.SelfHostedReasonFailedKeys
		return err
	}
	*obj = irsav1alpha1.SetupSelfHostedStatusReady(*obj, string(irsav1alpha1.SelfHostedReasonReady), "successfully setup resources for self-hosted")
	log.Info("the self-hosted resources have successfully set up")
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
