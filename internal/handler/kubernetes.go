package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (k *KubernetesHandler) DeleteAll(ctx context.Context) error {
	for _, obj := range k.objs {
		err := k.client.Delete(ctx, obj, DeleteOptions{
			DeletionPropagation: metav1.DeletePropagationBackground,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
