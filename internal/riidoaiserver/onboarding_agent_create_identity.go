package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

func onboardingAgentCreateIdentity(ctx context.Context, principal AuthorizationResult, workspaceID, fixtureID string, req CreateAgentConfigurationRequest) (string, string, error) {
	id, keyed := ctx.Value(onboardingAgentCreateRequestIDContextKey{}).(string)
	if !keyed {
		return "", "", nil
	}
	if len(id) < 1 || len(id) > 128 {
		return "", "", errors.New("request_id must be an ASCII token of 1 to 128 bytes")
	}
	for _, c := range id {
		valid := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == ':' || c == '-'
		if !valid {
			return "", "", errors.New("request_id must be an ASCII token of 1 to 128 bytes")
		}
	}
	if principal.PrincipalID == "" || workspaceID == "" {
		return "", "", ErrAuthorizationForbidden
	}
	scope, err := json.Marshal([]string{"onboarding-agent-create.v1", principal.PrincipalID, workspaceID, id})
	if err != nil {
		return "", "", err
	}
	payload, err := json.Marshal(struct {
		FixtureID string                          `json:"fixture_id"`
		Request   CreateAgentConfigurationRequest `json:"request"`
	}{fixtureID, req})
	if err != nil {
		return "", "", err
	}
	scopeHash, payloadHash := sha256.Sum256(scope), sha256.Sum256(payload)
	return hex.EncodeToString(scopeHash[:]), hex.EncodeToString(payloadHash[:]), nil
}
