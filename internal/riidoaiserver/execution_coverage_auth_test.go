package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCoverageDispatchRequiresExactAuthenticatedGeneration(t *testing.T) {
	f := newCoverageLifecycleFixture(t, "codex")
	f.link(t)
	f.assign(t)
	body, err := json.Marshal(PollRequest{DeviceID: f.enrollment.DeviceID, DaemonID: f.binding.DaemonID, RuntimeID: f.binding.RuntimeID, WaitMs: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		values []string
	}{
		{"missing", nil},
		{"wrong", []string{"untrusted-generation"}},
		{"duplicate", []string{f.enrollment.ExecutionCoverageGeneration, f.enrollment.ExecutionCoverageGeneration}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/agents/"+f.binding.AgentID+"/poll", bytes.NewReader(body))
			req.Header.Set(deviceIDHeader, f.enrollment.DeviceID)
			req.Header.Set(deviceSecretHeader, f.enrollment.DeviceSecret)
			for _, value := range tc.values {
				req.Header.Add(executionCoverageHeader, value)
			}
			response := httptest.NewRecorder()
			f.cli.handler.ServeHTTP(response, req)
			if response.Code != http.StatusForbidden {
				t.Fatalf("invalid generation dispatch status=%d", response.Code)
			}
			f.policy(t, "allowed")
		})
	}
	f.poll(t)
	f.policy(t, "blocked")
}
