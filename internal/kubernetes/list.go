package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// List lists all resources of the given kind.
func (c KubernetesClient) List(ctx context.Context, gvk schema.GroupVersionKind, listOpts ...client.ListOption) (*unstructured.UnstructuredList, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(gvk)
	err := c.client.List(ctx, list, listOpts...)
	if err != nil {
		return nil, err
	}

	return list, nil
}
