package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenericAssignmentCannotBypassRecoveryInventoryAuthorization(t *testing.T) {
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{PrincipalID: "user-1", Token: "token", Scopes: []string{"component-task:task:assign"}}})
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore()
	defer store.Close()
	server := NewServer(ServerConfig{Assignment: store, Authorizer: auth}).Handler()
	body := `{"component_id":"task","agent_id":"agent","runtime_provider":"codex","prompt":"perform requested fix","recovery_policy":{"max_switches":1,"fallbacks":[{"provider":"cursor","model_id":"cursor-auto"}]}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/component-tasks/task/assignment", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, req)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "scoped task assignment") {
		t.Fatalf("generic fallback bypass: status=%d body=%s", response.Code, response.Body.String())
	}
}
