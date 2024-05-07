package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesClient interface {
	Apply(ctx context.Context, obj client.Object) error
	Create(ctx context.Context, obj client.Object) error
	Get(ctx context.Context, obj client.Object) (*unstructured.Unstructured, error)
	Delete(ctx context.Context, obj client.Object, opts DeleteOptions) error
}

type StatusClient interface {
	PatchStatus(ctx context.Context, obj client.Object) error
}

// DeleteOptions contains options for delete requests.
type DeleteOptions struct {
	// DeletionPropagation decides how the garbage collector will handle the propagation.
	DeletionPropagation metav1.DeletionPropagation

	// Inclusions determines which in-cluster objects are subject to deletion
	// based on the labels.
	// A nil Inclusions map means all objects are subject to deletion
	Inclusions map[string]string
}
