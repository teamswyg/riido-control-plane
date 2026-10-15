package riidoaiserver

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRuntimeSnapshotFailuresObserveDecodeAndSyncSeparately(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	server := newRuntimeSnapshotHTTPTestServer(t, NewDevelopmentAIAgentClientStore())
	enrollment := enrollRuntimeSnapshotDevice(t, server)
	for _, body := range []string{`{"daemon_id":`, `{"daemon_id":"","runtimes":[]}`} {
		req := runtimeSnapshotRequest(body, enrollment)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("snapshot status=%d", response.Code)
		}
	}
	got := output.String()
	for _, want := range []string{`stage="runtime_snapshot_decode"`, `stage="runtime_snapshot_sync"`, `error_type="incomplete_body"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing stage %q in %q", want, got)
		}
	}
	if strings.Contains(got, enrollment.DeviceSecret) {
		t.Fatal("snapshot logging exposed device credential")
	}
}
