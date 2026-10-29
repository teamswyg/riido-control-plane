package riidoaiserver

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func normalizeRuntimeSnapshotVersion(runtime *RuntimeSnapshotRecord) string {
	providerVersion, err := normalizeRuntimeProviderVersion(runtime.ProviderVersion)
	if err != nil {
		// Version is optional display metadata. Legacy probes may send warnings
		// or control characters; never reject the device heartbeat for that.
		providerVersion = ""
		if runtime.HealthStatus == "" || runtime.HealthStatus == hostintegration.ProviderHealthHealthy {
			runtime.HealthStatus = hostintegration.ProviderHealthUnknown
			runtime.DiagnosticCode = hostintegration.ProviderDiagnosticVersionProbeFailed
			runtime.DiagnosticSummary = ""
		}
	}
	return providerVersion
}

func normalizeRuntimeProviderVersion(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if utf8.RuneCountInString(value) > 128 {
		return "", errors.New("must be 128 characters or fewer")
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return "", errors.New("must not contain control characters")
		}
	}
	return value, nil
}
