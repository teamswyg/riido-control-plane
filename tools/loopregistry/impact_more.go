package main

import (
	"fmt"
	"strings"
)

func claimsByID(claims []claimBinding) map[string]claimBinding {
	out := map[string]claimBinding{}
	for _, claim := range claims {
		out[claim.ID] = claim
	}
	return out
}

func claimSignature(claim claimBinding) string {
	parts := []string{claim.ID, claim.Statement, claim.Loop}
	parts = append(parts, prefixedValues("file", claim.Files)...)
	parts = append(parts, prefixedValues("verifier", claim.Verifiers)...)
	return strings.Join(parts, "\x00")
}

func verifyChangedClaimImpact(claim claimBinding, changed map[string]bool) (impactClaim, error) {
	record := impactClaim{ID: claim.ID}
	for _, path := range claim.Files {
		if changed[path] {
			record.ChangedBoundFiles = append(record.ChangedBoundFiles, path)
		}
	}
	if len(record.ChangedBoundFiles) == 0 {
		return record, fmt.Errorf("claim %s changed without a bound code/test file change", claim.ID)
	}
	return record, nil
}

func prefixedValues(prefix string, values []string) []string {
	out := []string{}
	for _, value := range sortedCopy(values) {
		out = append(out, prefix+":"+value)
	}
	return out
}
