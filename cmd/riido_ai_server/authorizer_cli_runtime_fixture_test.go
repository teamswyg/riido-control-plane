package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const (
	cliRuntimeWorkspace = "workspace-dev-riid"
	cliRuntimeTask      = "task-fixture"
	cliRuntimeResource  = "https://cp.example.test"
	cliRuntimeSecret    = "synthetic-local-legacy-secret"
	cliRuntimeIssuer    = "https://example.com"
	cliRuntimeThread    = "/v3/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-fixture/threads"
	cliRuntimeAgents    = "/v1/cli/workspaces/workspace-dev-riid/tasks/task-fixture/agents"
	cliRuntimeApp       = "/v2/client/workspaces/workspace-dev-riid/ai-agent/bootstrap"
	cliRuntimeWrite     = "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-fixture/assignment"
)

type cliRuntimeFixture struct {
	server                        *httptest.Server
	legacy, thread, agents, other atomic.Int32
	policyStatus                  atomic.Int32
	substitute                    atomic.Bool
}

func cliRuntimeTicket(agent bool) string {
	prefix := riidoaiserver.CLIThreadReadTokenPrefix
	if agent {
		prefix = riidoaiserver.CLIAgentReadTokenPrefix
	}
	return prefix + strings.Repeat("A", 43)
}

func newCLIRuntimeFixture(t *testing.T, dual bool) *cliRuntimeFixture {
	t.Helper()
	clearRiidoAIServerEnv(t)
	f := &cliRuntimeFixture{}
	f.policyStatus.Store(http.StatusOK)
	f.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/legacy":
			f.legacy.Add(1)
			var request struct {
				Token string `json:"bearer_token"`
			}
			if r.Method != http.MethodPost || json.NewDecoder(r.Body).Decode(&request) != nil || !cliRuntimeLegacyToken(request.Token) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"schema_version": riidoaiserver.ExternalAuthorizerResponseSchemaVersion, "allowed": true, "principal_id": "user-1"})
		case "/internal/cli/thread-read/authorize", "/internal/cli/agent-read/authorize":
			f.policy(t, w, r)
		default:
			f.other.Add(1)
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
	t.Cleanup(f.server.Close)
	// Exercise the real environment factories and HTTP clients. Only this local
	// TLS fixture is trusted; no transport callback substitutes for network I/O.
	previous := http.DefaultTransport
	transport := f.server.Client().Transport.(*http.Transport).Clone()
	localAddress := f.server.Listener.Addr().String()
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address == "example.com:443" {
			address = localAddress
		}
		if address != localAddress {
			return nil, fmt.Errorf("unexpected fixture destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}
	http.DefaultTransport = transport
	t.Cleanup(transport.CloseIdleConnections)
	t.Cleanup(func() { http.DefaultTransport = previous })
	t.Setenv(envExternalAuthzURL, f.server.URL+"/legacy")
	for _, pair := range [][2]string{{envCLIThreadReadEnabled, "true"}, {envCLIAgentReadEnabled, "true"}, {envCLIThreadReadMonolithOrigin, f.server.URL}, {envCLIAgentReadMonolithOrigin, f.server.URL}, {envCLIThreadReadResource, cliRuntimeResource}, {envCLIAgentReadResource, cliRuntimeResource}, {envCLIThreadReadServiceToken, strings.Repeat("t", 32)}, {envCLIAgentReadServiceToken, strings.Repeat("a", 32)}} {
		t.Setenv(pair[0], pair[1])
	}
	if dual {
		setAuthProfileEnv(t)
		t.Setenv(envAuthIssuer, cliRuntimeIssuer)
	}
	return f
}

func cliRuntimeLegacyToken(raw string) bool {
	if raw == "synthetic-opaque-app" {
		return true
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(cliRuntimeSecret))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	return err == nil && hmac.Equal(sig, mac.Sum(nil))
}

func (f *cliRuntimeFixture) policy(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	agent := r.URL.Path == "/internal/cli/agent-read/authorize"
	purpose, schema, service := "task_threads_read", "riido-cli-thread-read.v1", strings.Repeat("t", 32)
	if agent {
		f.agents.Add(1)
		purpose, schema, service = "task_agents_read", "riido-cli-agent-read.v1", strings.Repeat("a", 32)
	} else {
		f.thread.Add(1)
	}
	var body map[string]string
	if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer "+service || r.Header.Get("Cookie") != "" || json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 5 || body["ticket"] != cliRuntimeTicket(agent) || body["purpose"] != purpose || body["resource"] != cliRuntimeResource || body["workspace_id"] != cliRuntimeWorkspace || body["task_id"] != cliRuntimeTask {
		t.Error("dedicated request lost its contract binding")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if status := int(f.policyStatus.Load()); status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	task := cliRuntimeTask
	if f.substitute.Load() {
		task = "other-task"
	}
	now := time.Now().UTC()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"schema_version": schema, "allowed": true, "principal_id": "user-1", "roles": []string{}, "resource": cliRuntimeResource, "purpose": purpose, "workspace_id": cliRuntimeWorkspace, "task_id": task, "checked_at": now, "expires_at": now.Add(55 * time.Second)})
}

func (f *cliRuntimeFixture) runtime(t *testing.T) *httptest.Server {
	t.Helper()
	a, err := authorizerFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	config, err := runtimeCLIThreadRead(runtimeConfig{Authorizer: a})
	if err != nil {
		t.Fatal(err)
	}
	store := riidoaiserver.NewDevelopmentAIAgentClientStore()
	_, err = store.CreateAIAgentTaskAgentAssignment(t.Context(), riidoaiserver.AuthorizationResult{PrincipalID: "user-1", WorkspaceID: cliRuntimeWorkspace}, cliRuntimeTask, riidoaiserver.AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"})
	if err != nil {
		t.Fatal(err)
	}
	assignment := riidoaiserver.NewStore()
	t.Cleanup(assignment.Close)
	s := httptest.NewServer(newRuntimeHandler(config, assignment, store, nil, nil))
	t.Cleanup(s.Close)
	return s
}

func cliRuntimeRequest(t *testing.T, s *httptest.Server, method, path, token, extra string) (int, []byte) {
	t.Helper()
	r, err := http.NewRequestWithContext(t.Context(), method, s.URL+path, strings.NewReader(`{"agent_id":"agent-owned-codex"}`))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "Bearer "+token)
	if extra != "" {
		r.Header.Set("X-Riido-Ai-Agent-Token", extra)
	}
	response, err := s.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response.StatusCode, raw
}
