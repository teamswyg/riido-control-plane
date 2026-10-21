package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func accessBoundaryAuthorizer(t *testing.T) RequestAuthorizer {
	t.Helper()
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request externalAuthorizerRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Request.Resource != AuthorizationResourceAIAgentClient || request.Request.WorkspaceID == "" {
			t.Errorf("authorization scope=%+v", request.Request)
		}
		response := externalAuthorizerResponse{SchemaVersion: ExternalAuthorizerResponseSchemaVersion, Allowed: true}
		switch request.BearerToken {
		case "owner-token":
			response.PrincipalID = "user-2"
		case "other-token":
			response.PrincipalID = "user-1"
		case "admin-token":
			response.PrincipalID = "admin-1"
			response.Roles = []AgentCatalogRole{AgentCatalogRoleAdmin}
		default:
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(authServer.Close)
	authorizer, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: authServer.URL})
	if err != nil {
		t.Fatal(err)
	}
	return authorizer
}

func accessBoundaryRequest(t *testing.T, server http.Handler, method, path, token, body string, want int) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path+"?principal_id=user-2&roles=admin", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Principal-Id", "user-2")
	request.Header.Set("X-Roles", "admin")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, response.Code, want, response.Body.String())
	}
	return response
}
