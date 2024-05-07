package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func (c KubernetesClient) toUnstructured(obj client.Object) (*unstructured.Unstructured, error) {
	gvk, err := apiutil.GVKForObject(obj, c.client.Scheme())
	if err != nil {
		return nil, err
	}

	u := &unstructured.Unstructured{}
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	u.Object = unstructured
	u.SetGroupVersionKind(gvk)
	u.SetManagedFields(nil)
	return u, nil
}
