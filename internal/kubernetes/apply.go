package kubernetes

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c KubernetesClient) Apply(ctx context.Context, obj client.Object) error {
	opts := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner(c.owner.Field),
	}
	u, err := c.toUnstructured(obj)
	if err != nil {
		return err
	}
	err = c.client.Patch(ctx, u, client.Apply, opts...)
	if err != nil {
		return err
	}
	return nil
}
