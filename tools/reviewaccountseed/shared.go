package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func reviewProvisioning() (riidoaiserver.ReviewAccountProvisioning, error) {
	seed, err := riidoaiserver.LoadReviewAccountSeed()
	if err != nil {
		return riidoaiserver.ReviewAccountProvisioning{}, err
	}
	return riidoaiserver.ProvisionReviewAccount(seed, riidoaiserver.ReviewAccountProvisionInput{
		TokenSHA256: tokenHash(reviewToken()),
	})
}

func reviewToken() string {
	return strings.Join([]string{"review", "token"}, "-")
}

func tokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
