package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func normalizeRuntimeProviderHealth(runtime RuntimeSnapshotRecord) (
	hostintegration.ProviderHealthStatus,
	hostintegration.ProviderDiagnosticCode,
	string,
	error,
) {
	health := hostintegration.ProviderHealthStatus(strings.TrimSpace(string(runtime.HealthStatus)))
	code := hostintegration.ProviderDiagnosticCode(strings.TrimSpace(string(runtime.DiagnosticCode)))
	if health == "" {
		if runtime.Availability == RuntimeAvailabilityOffline {
			health = hostintegration.ProviderHealthUnavailable
		} else {
			health = hostintegration.ProviderHealthHealthy
		}
	}
	if !health.Valid() {
		return "", "", "", fmt.Errorf("health_status %q is invalid", health)
	}
	if code == "" {
		code = hostintegration.ProviderDiagnosticNone
	}
	if !code.Valid() {
		return "", "", "", fmt.Errorf("diagnostic_code %q is invalid", code)
	}
	if health == hostintegration.ProviderHealthHealthy {
		code = hostintegration.ProviderDiagnosticNone
	} else if health == hostintegration.ProviderHealthUnknown && code == hostintegration.ProviderDiagnosticNone {
		code = hostintegration.ProviderDiagnosticProbeFailed
	}
	summary := canonicalProviderDiagnosticSummary(code)
	incomingSummary := strings.TrimSpace(runtime.DiagnosticSummary)
	if incomingSummary != "" && incomingSummary != summary {
		return "", "", "", errors.New("diagnostic_summary must match the bounded diagnostic code")
	}
	return health, code, summary, nil
}

func canonicalProviderDiagnosticSummary(code hostintegration.ProviderDiagnosticCode) string {
	switch code {
	case hostintegration.ProviderDiagnosticExecutableMissing:
		return "provider executable is not available"
	case hostintegration.ProviderDiagnosticLoginRequired:
		return "provider login is required"
	case hostintegration.ProviderDiagnosticVersionUnsupported:
		return "provider version is not supported"
	case hostintegration.ProviderDiagnosticProbeFailed:
		return "provider probe did not complete"
	case hostintegration.ProviderDiagnosticRuntimeError:
		return "provider runtime reported an error"
	default:
		return ""
	}
}
