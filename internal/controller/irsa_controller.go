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
	"fmt"

	awsclient "github.com/kkb0318/irsa-manager/internal/aws"
	"github.com/kkb0318/irsa-manager/internal/handler"
	"github.com/kkb0318/irsa-manager/internal/issuer"
	"github.com/kkb0318/irsa-manager/internal/kubernetes"
	"github.com/kkb0318/irsa-manager/internal/manifests"
	"github.com/kkb0318/irsa-manager/internal/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	irsav1alpha1 "github.com/kkb0318/irsa-manager/api/v1alpha1"
)

// IRSAReconciler reconciles a IRSA object
type IRSAReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	AwsClient awsclient.AwsClient
}

//+kubebuilder:rbac:groups=irsa-manager.kkb0318.github.io,resources=irsas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=irsa-manager.kkb0318.github.io,resources=irsas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=irsa-manager.kkb0318.github.io,resources=irsas/finalizers,verbs=update
//+kubebuilder:rbac:groups=irsa-manager.kkb0318.github.io,resources=irsasetups,verbs=get;list
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *IRSAReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	obj := &irsav1alpha1.IRSA{}
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
		if err := r.Get(ctx, req.NamespacedName, &irsav1alpha1.IRSA{}); err != nil {
			return
		}
		statusHandler := handler.NewStatusHandler(kubeClient)
		if err := statusHandler.Patch(ctx, obj); err != nil {
			return
		}
	}()

	if !obj.DeletionTimestamp.IsZero() {
		err = r.reconcileDelete(ctx, obj, kubeClient)
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

func (r *IRSAReconciler) reconcileDelete(ctx context.Context, obj *irsav1alpha1.IRSA, kubeClient *kubernetes.KubernetesClient) error {
	if !obj.Spec.Cleanup {
		return nil
	}
	roleManager := awsclient.RoleManager{
		RoleName: obj.Spec.IamRole.Name,
		Policies: obj.Spec.IamPolicies,
	}
	err := r.AwsClient.IamClient().DeleteIRSARole(
		ctx,
		roleManager,
	)
	if err != nil {
		return err
	}
	deleted, err := cleanupKubernetesResources(ctx, kubeClient, obj.Spec.ServiceAccount.NamespacedNameList())
	*obj = irsav1alpha1.IRSAStatusSetServiceAccount(*obj, deleted)
	if err != nil {
		return err
	}
	return nil
}

func (r *IRSAReconciler) reconcile(ctx context.Context, obj *irsav1alpha1.IRSA, kubeClient *kubernetes.KubernetesClient) error {
	list, err := kubeClient.List(ctx, irsav1alpha1.GroupVersion.WithKind(irsav1alpha1.IRSASetupKind))
	if err != nil {
		return err
	}
	if len(list.Items) != 1 {
		return fmt.Errorf("there should be exactly one IRSASetup item")
	}
	irsaSetup := &irsav1alpha1.IRSASetup{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(list.Items[0].Object, irsaSetup)
	if err != nil {
		return fmt.Errorf("error converting to IRSASetup for %s: %v", list.Items[0].GetName(), err)
	}
	serviceAccount := obj.Spec.ServiceAccount
	issuerMeta, err := issuer.NewOIDCIssuerMeta(irsaSetup)
	if err != nil {
		return err
	}
	// e is set only when an error occurs in an external dependency process and is reflected in the CRs status
	var e error
	var reason irsav1alpha1.IRSAReason
	defer func() {
		if e != nil {
			*obj = irsav1alpha1.IRSAStatusNotReady(*obj, string(reason), e.Error())
		}
	}()

	accountId, err := r.AwsClient.StsClient().GetAccountId()
	if err != nil {
		e = err
		return err
	}
	roleManager := awsclient.RoleManager{
		RoleName:       obj.Spec.IamRole.Name,
		ServiceAccount: serviceAccount,
		Policies:       obj.Spec.IamPolicies,
		AccountId:      accountId,
	}
	err = r.AwsClient.IamClient().UpdateIRSARole(
		ctx,
		issuerMeta,
		roleManager,
	)
	if err != nil {
		e = err
		reason = irsav1alpha1.IRSAReasonFailedRoleUpdate
		return err
	}

	kubeHandler := handler.NewKubernetesHandler(kubeClient)
	for _, namespacedName := range serviceAccount.NamespacedNameList() {
		sa := manifests.NewServiceAccountBuilder().WithIRSAAnnotation(roleManager).Build(namespacedName)
		kubeHandler.Append(sa)
	}
	applied, err := kubeHandler.ApplyAll(ctx)
	*obj = irsav1alpha1.IRSAStatusSetServiceAccount(*obj, applied)
	if err != nil {
		e = err
		reason = irsav1alpha1.IRSAReasonFailedK8sApply
		return err
	}

	deleted, err := cleanupKubernetesResources(
		ctx,
		kubeClient,
		utils.DiffNamespacedNames(obj.Status.ServiceNamespacedNameList(), serviceAccount.NamespacedNameList()),
	)
	*obj = irsav1alpha1.IRSAStatusRemoveServiceAccount(*obj, deleted)
	if err != nil {
		e = err
		reason = irsav1alpha1.IRSAReasonFailedK8sCleanUp
		return err
	}
	*obj = irsav1alpha1.IRSAStatusReady(*obj, string(irsav1alpha1.IRSAReasonReady), "successfully setup resources")
	return nil
}

func cleanupKubernetesResources(ctx context.Context, client *kubernetes.KubernetesClient, nsNames []types.NamespacedName) ([]types.NamespacedName, error) {
	kubeHandler := handler.NewKubernetesHandler(client)
	for _, namespacedName := range nsNames {
		sa := manifests.NewServiceAccountBuilder().Build(namespacedName)
		kubeHandler.Append(sa)
	}
	deleted, err := kubeHandler.DeleteAll(ctx)
	if err != nil {
		return deleted, err
	}
	return deleted, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IRSAReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&irsav1alpha1.IRSA{}).
		Complete(r)
}
