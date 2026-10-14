package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func reviewAccountProvisioningFromEnv() (*riidoaiserver.ReviewAccountProvisioning, error) {
	tokenHash := strings.TrimSpace(os.Getenv(envReviewAccountTokenHash))
	if tokenHash == "" {
		return nil, nil
	}
	seed, err := riidoaiserver.LoadReviewAccountSeed()
	if err != nil {
		return nil, fmt.Errorf("%s load seed: %w", envReviewAccountTokenHash, err)
	}
	provisioning, err := riidoaiserver.ProvisionReviewAccount(seed, riidoaiserver.ReviewAccountProvisionInput{
		TokenSHA256: tokenHash,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envReviewAccountTokenHash, err)
	}
	return &provisioning, nil
}
