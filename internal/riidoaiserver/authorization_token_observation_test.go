package riidoaiserver

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuthorizationDenialLogsFiniteTokenHintsWithoutChangingDecision(t *testing.T) {
	const canary = "synthetic-private-auth-canary"
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d,"sub":%q,"iss":%q}`, time.Now().Add(-time.Hour).Unix(), canary, canary)))
	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/private-workspace/ai-agent/tasks/private-task/thread-stream-subscription", nil)
	req.Header.Set(aiAgentTokenHeader, "e30."+payload+".signature")
	req.Header.Set("Authorization", "Bearer different-synthetic-token")
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	response := httptest.NewRecorder()
	_, ok := (Server{config: ServerConfig{Authorizer: observationFailingAuthorizer{ErrAuthorizationUnauthenticated}}}).authorizeRequest(response, req, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient})
	if ok || response.Code != http.StatusUnauthorized {
		t.Fatal("diagnostics changed denial")
	}
	for _, want := range []string{`token_source="agent_header"`, `token_format="jwt"`, `expiry_hint="past_unverified"`, `token_header_relation="both_different"`} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing finite diagnostic %s", want)
		}
	}
	for _, forbidden := range []string{canary, payload, "signature", "different-synthetic-token", "private-workspace", "private-task"} {
		if strings.Contains(output.String(), forbidden) || strings.Contains(response.Body.String(), forbidden) {
			t.Fatal("private authentication input escaped")
		}
	}
}
