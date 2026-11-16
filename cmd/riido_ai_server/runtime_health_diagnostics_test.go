package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRuntimeHealthDiagnosticsDistinguishConnectionAndReadyProbe(t *testing.T) {
	logFile, err := os.CreateTemp(t.TempDir(), "runtime-stderr-")
	if err != nil {
		t.Fatal(err)
	}
	previousStderr := os.Stderr
	os.Stderr = logFile
	t.Cleanup(func() {
		os.Stderr = previousStderr
		_ = logFile.Close()
	})
	server := newRuntimeHTTPServers(runtimeConfig{Addr: "127.0.0.1:0"}, nil, nil, nil, nil)[0]
	live := httptest.NewUnstartedServer(server.Handler)
	live.Config.ConnState = server.ConnState
	live.Start()
	t.Cleanup(live.Close)
	readLog := func() string {
		t.Helper()
		data, err := os.ReadFile(logFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	request := func(method, path string, wantStatus int) {
		t.Helper()
		req, err := http.NewRequest(method, live.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Close = true
		req.Header.Set("Authorization", "Bearer must-not-log-this")
		response, err := live.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != wantStatus {
			t.Fatalf("%s %s: status=%d, body=%s", method, path, response.StatusCode, body)
		}
		if method == http.MethodGet && strings.HasPrefix(path, "/readyz") && string(body) != "{\"schema_version\":\"riido-ai-server.v1\",\"status\":\"ready\"}\n" {
			t.Fatalf("ready response changed: %q", body)
		}
	}

	request(http.MethodGet, "/healthz?private=must-not-log-this", http.StatusOK)
	request(http.MethodPost, "/readyz", http.StatusMethodNotAllowed)
	if got := readLog(); got != "riido_ai_server: http_connection_received\n" {
		t.Fatalf("before ready probe, diagnostics=%q; want only the first connection marker", got)
	}
	request(http.MethodGet, "/readyz?private=must-not-log-this", http.StatusOK)
	request(http.MethodGet, "/readyz", http.StatusOK)
	if got := readLog(); got != "riido_ai_server: http_connection_received\nriido_ai_server: ready_probe_received\n" {
		t.Fatalf("diagnostics=%q; want one constant marker for each boundary", got)
	}
}
