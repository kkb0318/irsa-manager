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
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	irsav1alpha1 "github.com/kkb0318/irsa-manager/api/v1alpha1"
	awsclient "github.com/kkb0318/irsa-manager/internal/client"
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
	"github.com/kkb0318/irsa-manager/internal/selfhosted/oidc"
)

// IRSASetupReconciler reconciles a IRSASetup object
type IRSASetupReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	AwsClient awsclient.AwsClient
}

//+kubebuilder:rbac:groups=irsa.kkb0318.github.io,resources=irsasetups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=irsa.kkb0318.github.io,resources=irsasetups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=irsa.kkb0318.github.io,resources=irsasetups/finalizers,verbs=update

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
	if err := r.reconcile(ctx, obj); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("successfully reconciled")
	return ctrl.Result{}, nil
}

func (r *IRSASetupReconciler) reconcile(ctx context.Context, obj *irsav1alpha1.IRSASetup) error {
	err := reconcileSelfhosted(ctx, obj, r.AwsClient)
	return err
}

func reconcileSelfhosted(ctx context.Context, obj *irsav1alpha1.IRSASetup, awsClient awsclient.AwsClient) error {
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
	err = selfhosted.Execute(ctx, factory)
	if err != nil {
		return err
	}
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
