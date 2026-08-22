package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRuntimeHTTPServersBuildsPrimaryServer(t *testing.T) {
	servers, err := newRuntimeHTTPServers(runtimeConfig{Addr: "127.0.0.1:0"}, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 {
		t.Fatalf("server count = %d, want 1", len(servers))
	}
	server := servers[0]
	if server.Addr != "127.0.0.1:0" {
		t.Fatalf("addr = %q", server.Addr)
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("read header timeout = %s", server.ReadHeaderTimeout)
	}
	assertRuntimeHTTPStatus(t, server.Handler, "/healthz", "ok")
	assertRuntimeHTTPStatus(t, server.Handler, "/readyz", "ready")
}

func TestNewRuntimeHTTPServersAddsPprofServer(t *testing.T) {
	servers, err := newRuntimeHTTPServers(runtimeConfig{
		Addr:      "127.0.0.1:0",
		PprofAddr: "127.0.0.1:6060",
	}, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 2 {
		t.Fatalf("server count = %d, want 2", len(servers))
	}
	if servers[1].Addr != "127.0.0.1:6060" {
		t.Fatalf("pprof addr = %q", servers[1].Addr)
	}
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	resp := httptest.NewRecorder()
	servers[1].Handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), "profile") {
		t.Fatalf("pprof status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestNewRuntimeHandlerPreservesWebCORSConfig(t *testing.T) {
	handler, err := newRuntimeHandler(runtimeConfig{
		WebAllowedOrigins: []string{"https://app.riido.io"},
	}, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://app.riido.io")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "https://app.riido.io" {
		t.Fatalf("access-control-allow-origin = %q", got)
	}
}

func TestRuntimeStartupFailsWhenGraphQLContractAdmissionFails(t *testing.T) {
	handler, err := newRuntimeHandlerWithGraphQL(runtimeConfig{}, nil, nil, nil, nil, nil, errors.New("owner hash drift"))
	if err == nil || handler != nil || !strings.Contains(err.Error(), "owner hash drift") {
		t.Fatalf("handler=%v err=%v", handler, err)
	}
	if handler, err := newRuntimeHandlerWithGraphQL(runtimeConfig{}, nil, nil, nil, nil, nil, nil); err == nil || handler != nil {
		 t.Fatalf("nil GraphQL receiver was admitted: handler=%v err=%v", handler, err)
	}
}

func assertRuntimeHTTPStatus(t *testing.T, handler http.Handler, path, want string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), want) {
		t.Fatalf("%s status=%d body=%s", path, resp.Code, resp.Body.String())
	}
}
