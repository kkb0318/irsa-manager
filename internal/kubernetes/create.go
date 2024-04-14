package kubernetes

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h KubernetesClient) Create(ctx context.Context, obj client.Object) error {
	opts := []client.CreateOption{
		client.FieldOwner(h.owner.Field),
	}
	u, err := h.toUnstructured(obj)
	if err != nil {
		return err
	}
	err = h.client.Create(ctx, u, opts...)
	if err != nil {
		return err
	}
	return nil
}
