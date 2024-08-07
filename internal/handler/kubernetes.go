package handler

import (
	"context"
	"log"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesHandler struct {
	client KubernetesClient
	objs   []client.Object
}

func NewKubernetesHandler(c KubernetesClient) *KubernetesHandler {
	return &KubernetesHandler{
		client: c,
		objs:   []client.Object{},
	}
}

func (k *KubernetesHandler) Append(obj client.Object) {
	k.objs = append(k.objs, obj)
}

// CreateAll creates the given objects (AlreadyExists errors are ignored)
func (k *KubernetesHandler) CreateAll(ctx context.Context) error {
	for _, obj := range k.objs {
		err := k.client.Create(ctx, obj)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return err
			}
			log.Printf("resource %s/%s already exists. skipped to create \n", obj.GetNamespace(), obj.GetName())
		}
	}
	return nil
}

func (k *KubernetesHandler) ApplyAll(ctx context.Context) ([]types.NamespacedName, error) {
	applied := []types.NamespacedName{}
	for _, obj := range k.objs {
		err := k.client.Apply(ctx, obj)
		if err != nil {
			return applied, err
		}
		applied = append(applied, client.ObjectKeyFromObject(obj))
	}
	return applied, nil
}

func (k *KubernetesHandler) DeleteAll(ctx context.Context) ([]types.NamespacedName, error) {
	deleted := []types.NamespacedName{}
	for _, obj := range k.objs {
		err := k.client.Delete(ctx, obj, DeleteOptions{
			DeletionPropagation: metav1.DeletePropagationBackground,
		})
		if err != nil {
			return deleted, err
		}
		deleted = append(deleted, client.ObjectKeyFromObject(obj))
	}
	return deleted, nil
}
