package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type externalAuthorizerHarness struct {
	*httptest.Server
	calls atomic.Int32
}

func newExternalAuthorizerHarness(t *testing.T) *externalAuthorizerHarness {
	t.Helper()
	return newExternalAuthorizerHarnessForToken(t, "external-token")
}

func newExternalAuthorizerHarnessForToken(t *testing.T, expectedToken string) *externalAuthorizerHarness {
	t.Helper()
	h := &externalAuthorizerHarness{}
	h.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.calls.Add(1)
		if got := r.Header.Get(riidoaiserver.ExternalAuthorizerAPIKeyHeader); got != "internal-key" {
			t.Fatalf("external authorizer api key header = %q", got)
		}
		assertExternalAuthorizerRequest(t, r, expectedToken)
		_ = json.NewEncoder(w).Encode(struct {
			SchemaVersion string `json:"schema_version"`
			Allowed       bool   `json:"allowed"`
			PrincipalID   string `json:"principal_id"`
		}{riidoaiserver.ExternalAuthorizerResponseSchemaVersion, true, "external-user"})
	}))
	return h
}

func (h *externalAuthorizerHarness) Calls() int32 {
	return h.calls.Load()
}
