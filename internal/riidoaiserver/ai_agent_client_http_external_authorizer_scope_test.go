package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientExternalAuthorizerRequiresWorkspaceScopedRoute(t *testing.T) {
	var calls int
	var got externalAuthorizerRequest
	authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&got); err != nil {
			t.Fatalf("decode authorizer request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(externalAuthorizerResponse{
			SchemaVersion: ExternalAuthorizerResponseSchemaVersion,
			Allowed:       true,
			PrincipalID:   "user-1",
		})
	}))
	defer authorizerServer.Close()
	authorizer, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{
		Endpoint: authorizerServer.URL,
	})
	if err != nil {
		t.Fatalf("NewExternalHTTPAuthorizer: %v", err)
	}
	server := NewServer(ServerConfig{
		AIAgentClient: NewDevelopmentAIAgentClientStore(),
		Authorizer:    authorizer,
	}).Handler()

	v1Req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	v1Req.Header.Set(aiAgentTokenHeader, "external-token")
	v1Resp := httptest.NewRecorder()
	server.ServeHTTP(v1Resp, v1Req)
	if v1Resp.Code != http.StatusForbidden {
		t.Fatalf("v1 status=%d body=%s", v1Resp.Code, v1Resp.Body.String())
	}
	if calls != 0 {
		t.Fatalf("workspace-less v1 request reached authorizer %d times", calls)
	}

	v2Req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-1/ai-agent/bootstrap", nil)
	v2Req.Header.Set(aiAgentTokenHeader, "external-token")
	v2Resp := httptest.NewRecorder()
	server.ServeHTTP(v2Resp, v2Req)
	if v2Resp.Code != http.StatusOK {
		t.Fatalf("v2 status=%d body=%s", v2Resp.Code, v2Resp.Body.String())
	}
	if calls != 1 {
		t.Fatalf("authorizer calls=%d, want 1", calls)
	}
	if got.Request.WorkspaceID != "workspace-1" || got.Request.Resource != AuthorizationResourceAIAgentClient {
		t.Fatalf("authorizer request = %+v", got)
	}
}
