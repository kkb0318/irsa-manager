package kubernetes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kkb0318/irsa-manager/internal/handler"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Delete deletes the given object (not found errors are ignored).
func (h *KubernetesClient) Delete(ctx context.Context, obj client.Object, opts handler.DeleteOptions) error {
	u, err := h.toUnstructured(obj)
	if err != nil {
		return err
	}
	existingObject := &unstructured.Unstructured{}
	existingObject.SetGroupVersionKind(u.GroupVersionKind())
	err = h.client.Get(ctx, client.ObjectKeyFromObject(u), existingObject)
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete: %w", err)
		}
		return nil // already deleted
	}

	sel, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: opts.Inclusions})
	if err != nil {
		return fmt.Errorf("label selector failed: %w", err)
	}

	if !sel.Matches(labels.Set(existingObject.GetLabels())) {
		return nil
	}

	if err := h.client.Delete(ctx, existingObject, client.PropagationPolicy(opts.DeletionPropagation)); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}
