package riidoaiserver

import "fmt"

func (s *DevelopmentAIAgentClientStore) enrollmentDeviceIDLocked(enrollment normalizedDeviceEnrollment) (string, error) {
	switch {
	case enrollment.machineID != "":
		deviceID := deviceIDForMachine(enrollment.machineID)
		existing, ok := s.deviceCredentials[deviceID]
		if ok && existing.ownerPrincipalID != enrollment.principalID {
			return "", ErrAuthorizationForbidden
		}
		return deviceID, nil
	case enrollment.deviceID != "":
		existing, ok := s.deviceCredentials[enrollment.deviceID]
		if !ok ||
			existing.ownerPrincipalID != enrollment.principalID ||
			existing.workspaceID != enrollment.workspaceID {
			return "", ErrAuthorizationForbidden
		}
		return enrollment.deviceID, nil
	default:
		s.nextDeviceCredentialSeq++
		return fmt.Sprintf("device-enrolled-%06d", s.nextDeviceCredentialSeq), nil
	}
}
