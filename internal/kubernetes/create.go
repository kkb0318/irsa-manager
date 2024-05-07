package kubernetes

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c KubernetesClient) Create(ctx context.Context, obj client.Object) error {
	opts := []client.CreateOption{
		client.FieldOwner(c.owner.Field),
	}
	u, err := c.toUnstructured(obj)
	if err != nil {
		return err
	}
	err = c.client.Create(ctx, u, opts...)
	if err != nil {
		return err
	}
	return nil
}
