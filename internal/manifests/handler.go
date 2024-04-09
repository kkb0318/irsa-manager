package manifests

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Handler interface {
	Apply(ctx context.Context, obj *unstructured.Unstructured) error
	Delete(ctx context.Context, obj *unstructured.Unstructured) error
}
