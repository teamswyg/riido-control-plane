package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestCLIInstallationServiceConfigurationUsesExactExistingCredentialScope(t *testing.T) {
	clearRiidoAIServerEnv(t)
	if cfg, err := cliInstallationServiceFromEnv(); err != nil || cfg.Resource != "" {
		t.Fatal("absent optional config broke legacy startup")
	}
	t.Setenv(envCLIInstallationResource, "https://cli.fixture.invalid")
	t.Setenv(envAuthzTokensJSON, `[{"principal_id":"service:riido-monolith","token":"synthetic-service-token","scopes":["cli-installation:link"]}]`)
	cfg, err := cliInstallationServiceFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	handler := riidoaiserver.NewServer(riidoaiserver.ServerConfig{CLIInstallationService: cfg}).Handler()
	for path, want := range map[string]int{"link-intents": http.StatusBadRequest, "policy": http.StatusForbidden} {
		r := httptest.NewRequest(http.MethodPost, "/v1/internal/cli-installations/"+path, strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer synthetic-service-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s scope=%d want=%d", path, w.Code, want)
		}
	}
	t.Setenv(envAuthzTokensJSON, `[{"principal_id":"user","token":"synthetic-user-token","scopes":["riido:*"]}]`)
	if _, err = cliInstallationServiceFromEnv(); err == nil {
		t.Fatal("user credential became service authority")
	}
}
