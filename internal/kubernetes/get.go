package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Get gets the given object.
func (c KubernetesClient) Get(ctx context.Context, obj client.Object) (*unstructured.Unstructured, error) {
	u, err := c.toUnstructured(obj)
	if err != nil {
		return nil, err
	}
	existingObj := &unstructured.Unstructured{}
	existingObj.SetGroupVersionKind(u.GroupVersionKind())
	err = c.client.Get(ctx, client.ObjectKeyFromObject(u), existingObj)
	if err != nil {
		return nil, err
	}
	return existingObj, nil
}
