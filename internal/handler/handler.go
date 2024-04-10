package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesClient interface {
	Apply(ctx context.Context, obj client.Object) error
	Delete(ctx context.Context, obj *unstructured.Unstructured, opts DeleteOptions) error
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
