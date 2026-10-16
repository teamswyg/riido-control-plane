package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type unavailablePresenceStore struct {
	*DevelopmentAIAgentClientStore
	cause error
}

func (s unavailablePresenceStore) DesktopPresenceLastSeenAt(context.Context, AuthorizationResult) (time.Time, error) {
	return time.Time{}, s.cause
}

func TestHTTPDesktopPresenceDependencyErrorsAreSafeAndRetryable(t *testing.T) {
	const private = "private-token-and-upstream-body-must-not-escape"
	for _, tc := range []struct {
		name   string
		cause  error
		status int
	}{
		{"dependency", errors.New(private), 503},
		{"cancelled", fmt.Errorf("%s: %w", private, context.Canceled), 408},
		{"deadline", fmt.Errorf("%s: %w", private, context.DeadlineExceeded), 504}, {"forbidden", ErrAuthorizationForbidden, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := unavailablePresenceStore{NewDevelopmentAIAgentClientStore(), tc.cause}
			handler := NewServer(ServerConfig{AIAgentClient: store, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "web-user")}).Handler()
			req := httptest.NewRequest(http.MethodGet, desktopPresenceTestPath, nil)
			req.Header.Set(aiAgentTokenHeader, "ai-agent-token")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			if resp.Code != tc.status || strings.Contains(resp.Body.String(), private) {
				t.Fatalf("unsafe or misclassified presence response: status=%d", resp.Code)
			}
		})
	}
}

func TestDesktopPresenceValidatesIdentityAndLegacyTime(t *testing.T) {
	now := time.Now().UTC()
	grant := DeviceConnectionGrant{DeviceID: " device ", PrincipalID: " owner ", WorkspaceID: " workspace ", ConnectedAt: now}
	seen, err := desktopPresenceFromGrants([]DeviceConnectionGrant{grant}, "owner")
	if err != nil || !seen.Equal(now) {
		t.Fatalf("legacy connection time missing: %v", err)
	}
	for _, field := range []string{"device", "principal", "workspace"} {
		invalid := grant
		switch field {
		case "device":
			invalid.DeviceID = " "
		case "principal":
			invalid.PrincipalID = " "
		case "workspace":
			invalid.WorkspaceID = " "
		}
		if _, err := desktopPresenceFromGrants([]DeviceConnectionGrant{invalid}, "owner"); err == nil {
			t.Fatalf("invalid %s identity accepted", field)
		}
	}
}
