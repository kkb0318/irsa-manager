package kubernetes

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c KubernetesClient) PatchStatus(ctx context.Context, obj client.Object) error {
	opts := &client.SubResourcePatchOptions{
		PatchOptions: client.PatchOptions{
			FieldManager: c.owner.Field,
		},
	}
	u, err := c.toUnstructured(obj)
	if err != nil {
		return err
	}

	return c.client.Status().Patch(ctx, u, client.Apply, opts)
}
