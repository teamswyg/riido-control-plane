package main

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func applyReviewAccountProvisioning(ctx context.Context, store *riidoaiserver.Store, config runtimeConfig) error {
	if config.ReviewProvision == nil {
		return nil
	}
	if err := store.ApplyReviewAccountProvisioning(ctx, *config.ReviewProvision); err != nil {
		return fmt.Errorf("apply review account provisioning: %w", err)
	}
	return nil
}
