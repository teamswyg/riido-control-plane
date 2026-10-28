package riidoaiserver

import (
	"errors"
	"strings"
)

const coverageLocalTermination = "local_invocations_quiescent"

func coverageTerminationScope(a Assignment) string {
	if !coverageLocalProvider(a.RuntimeProvider) {
		return "unobserved"
	}
	policy, err := a.RecoveryPolicy.Normalized()
	if err != nil {
		return "unobserved"
	}
	for _, fallback := range policy.Fallbacks {
		if !coverageLocalProvider(fallback.Provider) {
			return "unobserved"
		}
	}
	return coverageLocalTermination
}

func coverageLocalProvider(provider string) bool {
	return provider == "claude" || provider == "codex" || provider == "cursor"
}

func validateClaimCoverage(claim executionClaim) error {
	if claim.CoverageGeneration == "" {
		if claim.CredentialGeneration != "" || claim.RuntimeProvider != "" || claim.TerminationScope != "" {
			return errors.New("partial execution coverage provenance")
		}
		return nil
	}
	if !strings.HasPrefix(claim.DeviceID, executionCoveragePrefix) || !cliInstallationID(claim.CoverageGeneration) || !cliInstallationID(claim.CredentialGeneration) || !cliInstallationID(claim.RuntimeProvider) ||
		(claim.TerminationScope != coverageLocalTermination && claim.TerminationScope != "unobserved") {
		return errors.New("invalid execution coverage provenance")
	}
	return nil
}
