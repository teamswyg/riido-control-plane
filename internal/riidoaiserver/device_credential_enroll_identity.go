package riidoaiserver

import "fmt"

func (s *DevelopmentAIAgentClientStore) enrollmentDeviceIDLocked(enrollment normalizedDeviceEnrollment) (string, error) {
	switch {
	case enrollment.deviceID != "":
		existing, ok := s.deviceCredentials[enrollment.deviceID]
		if !ok ||
			existing.ownerPrincipalID != enrollment.principalID ||
			existing.workspaceID != enrollment.workspaceID {
			return "", ErrAuthorizationForbidden
		}
		return enrollment.deviceID, nil
	case enrollment.machineID != "":
		for deviceID, existing := range s.deviceCredentials {
			if existing.machineID != enrollment.machineID {
				continue
			}
			if existing.ownerPrincipalID != enrollment.principalID || (enrollment.deviceID != "" && enrollment.deviceID != deviceID) {
				return "", ErrAuthorizationForbidden
			}
			return deviceID, nil
		}
		deviceID := deviceIDForMachine(enrollment.machineID)
		existing, ok := s.deviceCredentials[deviceID]
		if ok && existing.ownerPrincipalID != enrollment.principalID {
			return "", ErrAuthorizationForbidden
		}
		if !ok && enrollment.coverageProtocol != "" {
			if enrollment.deviceID != "" {
				return "", ErrAuthorizationForbidden
			}
			return newCoverageIdentity(executionCoveragePrefix)
		}
		return deviceID, nil
	default:
		if enrollment.coverageProtocol != "" {
			return newCoverageIdentity(executionCoveragePrefix)
		}
		s.nextDeviceCredentialSeq++
		return fmt.Sprintf("device-enrolled-%06d", s.nextDeviceCredentialSeq), nil
	}
}
