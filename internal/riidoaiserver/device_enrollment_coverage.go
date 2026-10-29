package riidoaiserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
)

type executionCoverageEnrollmentStore interface {
	verifyEnrollmentCoverage(context.Context, EnrollDeviceResponse) error
}

// ConfigureExecutionCoverageEnrollment changes only admission of new namespaces.
// Existing covered credentials always retain their atomic rotation requirements.
func (s *DevelopmentAIAgentClientStore) ConfigureExecutionCoverageEnrollment(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executionCoverageEnrollment = enabled
}

func newCoverageIdentity(prefix string) (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(raw[:]), nil
}

func validateCredentialCoverage(c AIAgentClientDeviceCredentialSnapshot) error {
	covered := strings.HasPrefix(c.DeviceID, executionCoveragePrefix)
	if covered && cliInstallationID(c.ExecutionCoverageGeneration) && cliInstallationID(c.CredentialGeneration) {
		return nil
	}
	if !covered && c.ExecutionCoverageGeneration == "" && c.CredentialGeneration == "" {
		return nil
	}
	return errCLIInstallationUnavailable
}

func (s *DevelopmentAIAgentClientStore) enrollmentCoverageLocked(deviceID string, enrollment normalizedDeviceEnrollment) (string, string, error) {
	existing, existed := s.deviceCredentials[deviceID]
	if existed && existing.coverageGeneration != "" {
		if !s.executionCoverageAtomic || !strings.HasPrefix(deviceID, executionCoveragePrefix) || existing.credentialGeneration == "" {
			return "", "", errCLIInstallationUnavailable
		}
		credential, err := newCoverageIdentity("cred_")
		return existing.coverageGeneration, credential, err
	}
	if existed { // A protocol upgrade must never erase legacy execution history.
		if strings.HasPrefix(deviceID, executionCoveragePrefix) {
			return "", "", errCLIInstallationUnavailable
		}
		return "", "", nil
	}
	if enrollment.coverageProtocol == "" {
		return "", "", nil
	}
	if !s.executionCoverageEnrollment || !s.executionCoverageAtomic {
		return "", "", errCLIInstallationUnavailable
	}
	generation, err := newCoverageIdentity("cov_")
	if err != nil {
		return "", "", err
	}
	credential, err := newCoverageIdentity("cred_")
	return generation, credential, err
}
