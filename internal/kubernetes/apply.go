package kubernetes

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h KubernetesClient) Apply(ctx context.Context, obj client.Object) error {
	opts := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner(h.owner.Field),
	}
	u, err := h.toUnstructured(obj)
	if err != nil {
		return err
	}
	err = h.client.Patch(ctx, u, client.Apply, opts...)
	if err != nil {
		return err
	}
	return nil
}

func (h KubernetesClient) PatchStatus(ctx context.Context, obj client.Object) error {
	opts := &client.SubResourcePatchOptions{
		PatchOptions: client.PatchOptions{
			FieldManager: h.owner.Field,
		},
	}
	u, err := h.toUnstructured(obj)
	if err != nil {
		return err
	}

	return h.client.Status().Patch(ctx, u, client.Apply, opts)
}
