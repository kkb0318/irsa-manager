package manifests

import "context"

type Manifest interface {
	Apply(ctx context.Context, handler Handler) error
	Delete(ctx context.Context, handler Handler) error
}
