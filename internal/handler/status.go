package handler

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatusHandler struct {
	client StatusClient
}

func NewStatusHandler(c StatusClient) *StatusHandler {
	return &StatusHandler{
		client: c,
	}
}

func (s *StatusHandler) Patch(ctx context.Context, obj client.Object) error {
	return s.client.PatchStatus(ctx, obj)
}
