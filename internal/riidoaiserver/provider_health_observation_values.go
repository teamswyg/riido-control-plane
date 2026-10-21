package riidoaiserver

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func providerHealthDeviceRef(deviceID string) string {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return "unknown"
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(deviceID)))[:12]
}

func providerDiagnosticStage(code hostintegration.ProviderDiagnosticCode) string {
	switch code {
	case hostintegration.ProviderDiagnosticExecutableMissing:
		return "executable"
	case hostintegration.ProviderDiagnosticLoginRequired, hostintegration.ProviderDiagnosticAuthProbeFailed:
		return "authentication"
	case hostintegration.ProviderDiagnosticVersionUnsupported, hostintegration.ProviderDiagnosticVersionProbeFailed:
		return "version"
	case hostintegration.ProviderDiagnosticCapabilityProbeFailed:
		return "capability"
	case hostintegration.ProviderDiagnosticRuntimeError:
		return "runtime"
	case hostintegration.ProviderDiagnosticProbeFailed:
		return "probe"
	default:
		return "none"
	}
}
