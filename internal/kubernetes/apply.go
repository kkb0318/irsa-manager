package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Handler struct {
	cleanup bool
	client  client.Client
	owner   Owner
}

// NewHelper returns an initialized Helper.
func NewHandler(c client.Client, owner Owner, cleanup bool) (*Handler, error) {
	return &Handler{
		cleanup: cleanup,
		client:  c,
		owner:   owner,
	}, nil
}

func (h Handler) Apply(ctx context.Context, obj client.Object) error {
	opts := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner(h.owner.Field),
	}
	gvk, err := apiutil.GVKForObject(obj, h.client.Scheme())
	if err != nil {
		return err
	}

	u := &unstructured.Unstructured{}
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	u.Object = unstructured
	u.SetGroupVersionKind(gvk)
	u.SetManagedFields(nil)
	err = h.client.Patch(ctx, u, client.Apply, opts...)
	if err != nil {
		return err
	}
	return nil
}

func (h Handler) PatchStatus(ctx context.Context, obj client.Object) error {
	opts := &client.SubResourcePatchOptions{
		PatchOptions: client.PatchOptions{
			FieldManager: h.owner.Field,
		},
	}
	gvk, err := apiutil.GVKForObject(obj, h.client.Scheme())
	if err != nil {
		return err
	}

	u := &unstructured.Unstructured{}
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	u.Object = unstructured
	u.SetGroupVersionKind(gvk)
	u.SetManagedFields(nil)
	return h.client.Status().Patch(ctx, u, client.Apply, opts)
}
