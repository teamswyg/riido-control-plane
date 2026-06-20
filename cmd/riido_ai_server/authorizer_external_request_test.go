package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func assertExternalAuthorizerRequest(t *testing.T, r *http.Request) {
	t.Helper()
	var req struct {
		SchemaVersion string `json:"schema_version"`
		BearerToken   string `json:"bearer_token"`
		Request       struct {
			Resource string `json:"resource"`
			Action   string `json:"action"`
		} `json:"request"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if req.SchemaVersion != riidoaiserver.ExternalAuthorizerRequestSchemaVersion ||
		req.BearerToken != "external-token" ||
		req.Request.Resource != "metrics" ||
		req.Request.Action != "read" {
		t.Fatalf("external request = %+v", req)
	}
}
