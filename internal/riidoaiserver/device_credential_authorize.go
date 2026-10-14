package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) AuthorizeDeviceCredential(ctx context.Context, deviceID, deviceSecret string, req AuthorizationRequest) (AuthorizationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthorizationResult{}, err
	}
	deviceID = strings.TrimSpace(deviceID)
	deviceSecret = strings.TrimSpace(deviceSecret)
	if deviceID == "" || deviceSecret == "" {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	if req.Resource != AuthorizationResourceAgent {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	s.mu.Lock()
	record, ok := s.deviceCredentials[deviceID]
	s.mu.Unlock()
	if !ok {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	got := sha256.Sum256([]byte(deviceSecret))
	if subtle.ConstantTimeCompare(got[:], record.secretHash[:]) != 1 {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	workspaceID := strings.TrimSpace(req.WorkspaceID)
	if workspaceID == "" {
		workspaceID = record.workspaceID
	}
	return AuthorizationResult{
		PrincipalID: record.ownerPrincipalID,
		WorkspaceID: workspaceID,
	}, nil
}
