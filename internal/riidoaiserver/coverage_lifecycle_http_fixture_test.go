package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func (f *coverageLifecycleFixture) post(t *testing.T, path string, body any, mode, lease string) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	switch mode {
	case "device":
		req.Header.Set(deviceIDHeader, f.enrollment.DeviceID)
		req.Header.Set(deviceSecretHeader, f.enrollment.DeviceSecret)
		req.Header.Set(executionCoverageHeader, f.enrollment.ExecutionCoverageGeneration)
	case "service":
		req.Header.Set("Authorization", "Bearer service-fixture-token")
	case "human":
		req.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	case "bearer":
		req.Header.Set("Authorization", "Bearer ai-agent-token")
	default:
		t.Fatal("unknown fixture auth mode")
	}
	if lease != "" {
		req.Header.Set("X-Riido-Assignment-Lease-Token", lease)
	}
	recorder := httptest.NewRecorder()
	f.cli.handler.ServeHTTP(recorder, req)
	return recorder
}
