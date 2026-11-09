package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func (f *bootstrapFixture) handler(t *testing.T, scope string) http.Handler {
	t.Helper()
	authority, err := NewCLIInstallationServiceConfig(f.request.Resource, []StaticTokenCredential{{PrincipalID: cliInstallationServicePrincipal, Token: "bootstrap-fixture-service", Scopes: []string{scope}}})
	if err != nil {
		t.Fatal(err)
	}
	actor := NewStore()
	t.Cleanup(actor.Close)
	return NewServer(ServerConfig{Assignment: actor, AIAgentClient: f.store, CLIInstallationService: authority}).Handler()
}

func bootstrapHTTP(t *testing.T, handler http.Handler, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
