package handler

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesHandler struct {
	client KubernetesClient
	objs   []client.Object
}

func NewKubernetesHandler(k KubernetesClient) *KubernetesHandler {
	return &KubernetesHandler{
		client: k,
		objs:   []client.Object{},
	}
}

func (k *KubernetesHandler) Append(obj client.Object) {
	k.objs = append(k.objs, obj)
}

func (k *KubernetesHandler) ApplyAll(ctx context.Context) error {
	for _, obj := range k.objs {
		err := k.client.Apply(ctx, obj)
		if err != nil {
			return err
		}
	}
	return nil
}
