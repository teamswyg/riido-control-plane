package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestAuthorizerFromEnvCLIActualRuntimeBoundaries(t *testing.T) {
	for _, dual := range []bool{false, true} {
		name := "legacy-only"
		if dual {
			name = "dual-auth"
		}
		t.Run(name, func(t *testing.T) {
			f := newCLIRuntimeFixture(t, dual)
			s := f.runtime(t)
			rawCLI := signedCLIBoundaryFixture(`{"id":"user-1","clientId":"riido-cli","cliGrantId":"synthetic-grant","scopes":["read"],"iss":"https://monolith.example.test","exp":4102444800}`, []byte(cliRuntimeSecret))
			for _, tc := range []struct{ name, method, path, token, extra string }{
				{"raw CLI read", http.MethodGet, cliRuntimeApp, rawCLI, ""},
				{"raw CLI create", http.MethodPost, cliRuntimeWrite, rawCLI, ""},
				{"raw CLI thread", http.MethodGet, cliRuntimeThread, rawCLI, ""},
				{"raw CLI agent", http.MethodGet, cliRuntimeAgents, rawCLI, ""},
				{"thread to agent", http.MethodGet, cliRuntimeAgents, cliRuntimeTicket(false), ""},
				{"agent to thread", http.MethodGet, cliRuntimeThread, cliRuntimeTicket(true), ""},
				{"thread to write", http.MethodPost, cliRuntimeWrite, cliRuntimeTicket(false), ""},
				{"agent to write", http.MethodPost, cliRuntimeWrite, cliRuntimeTicket(true), ""},
				{"thread to app", http.MethodGet, cliRuntimeApp, cliRuntimeTicket(false), ""},
				{"agent to app", http.MethodGet, cliRuntimeApp, cliRuntimeTicket(true), ""},
				{"mixed app header", http.MethodGet, cliRuntimeThread, cliRuntimeTicket(false), "synthetic-opaque-app"},
				{"mixed reverse header", http.MethodGet, cliRuntimeThread, "synthetic-opaque-app", cliRuntimeTicket(false)},
				{"mixed ticket purposes", http.MethodGet, cliRuntimeThread, cliRuntimeTicket(false), cliRuntimeTicket(true)},
			} {
				t.Run(tc.name, func(t *testing.T) {
					if status, _ := cliRuntimeRequest(t, s, tc.method, tc.path, tc.token, tc.extra); status != http.StatusForbidden {
						t.Fatal("boundary accepted credential", status)
					}
					if f.legacy.Load() != 0 || f.thread.Load() != 0 || f.agents.Load() != 0 || f.other.Load() != 0 {
						t.Fatal("boundary contacted an authorization dependency")
					}
				})
			}
			// Positive controls use the same real server composition and HTTP PDP.
			for _, token := range []string{"synthetic-opaque-app", signedCLIBoundaryFixture(`{"id":"user-1","exp":4102444800}`, []byte(cliRuntimeSecret))} {
				if status, _ := cliRuntimeRequest(t, s, http.MethodGet, cliRuntimeApp, token, ""); status != http.StatusOK {
					t.Fatal("app compatibility failed", status)
				}
			}
			if f.legacy.Load() != 2 {
				t.Fatal("app positive controls missed legacy PDP")
			}
			status, raw := cliRuntimeRequest(t, s, http.MethodGet, cliRuntimeThread, cliRuntimeTicket(false), "")
			var history riidoaiserver.AIAgentTaskThreadHistoryCollectionResponse
			if status != http.StatusOK || json.Unmarshal(raw, &history) != nil || len(history.Threads) != 1 || history.Threads[0].ConversationID == "" {
				t.Fatal("thread projection failed", status)
			}
			status, raw = cliRuntimeRequest(t, s, http.MethodGet, cliRuntimeAgents, cliRuntimeTicket(true), "")
			var directory riidoaiserver.AgentDirectoryPage
			if status != http.StatusOK || json.Unmarshal(raw, &directory) != nil || directory.Total != 3 || len(directory.Agents) != 3 || directory.PolicyChecked {
				t.Fatal("agent projection failed", status)
			}
			for _, agent := range directory.Agents {
				if agent.AgentID == "agent-private-cursor" || agent.IsCurrentAgent {
					t.Fatal("agent projection exposed another owner or invented a current run")
				}
			}
			if f.legacy.Load() != 2 || f.thread.Load() != 1 || f.agents.Load() != 1 || f.other.Load() != 0 {
				t.Fatal("dedicated reader used wrong authority")
			}
			if dual {
				claims, _ := json.Marshal(map[string]string{"iss": cliRuntimeIssuer})
				invalidV2 := signedCLIBoundaryFixture(string(claims), []byte(cliRuntimeSecret))
				if status, _ := cliRuntimeRequest(t, s, http.MethodGet, cliRuntimeApp, invalidV2, ""); status != http.StatusUnauthorized {
					t.Fatal("Auth-v2 invalid token downgraded", status)
				}
				if f.legacy.Load() != 2 || f.other.Load() != 0 {
					t.Fatal("Auth-v2 rejection fell back")
				}
			}
			// Current revocation is checked again using the same runtime instance.
			f.policyStatus.Store(http.StatusForbidden)
			for _, target := range []struct {
				path  string
				agent bool
			}{{cliRuntimeThread, false}, {cliRuntimeAgents, true}} {
				if status, _ := cliRuntimeRequest(t, s, http.MethodGet, target.path, cliRuntimeTicket(target.agent), ""); status != http.StatusForbidden {
					t.Fatal("revoked read reused authority", status)
				}
			}
			if f.legacy.Load() != 2 || f.thread.Load() != 2 || f.agents.Load() != 2 || f.other.Load() != 0 {
				t.Fatal("revoked reader used fallback")
			}
		})
	}
}

func TestAuthorizerFromEnvCLIActualRuntimeReaderFailures(t *testing.T) {
	for _, mode := range []string{"disabled", "unavailable", "substituted-task"} {
		t.Run(mode, func(t *testing.T) {
			f := newCLIRuntimeFixture(t, true)
			if mode == "disabled" {
				t.Setenv(envCLIThreadReadEnabled, "false")
				t.Setenv(envCLIAgentReadEnabled, "false")
			}
			if mode == "unavailable" {
				f.policyStatus.Store(http.StatusServiceUnavailable)
			}
			if mode == "substituted-task" {
				f.substitute.Store(true)
			}
			wantStatus := http.StatusServiceUnavailable
			if mode == "substituted-task" {
				wantStatus = http.StatusForbidden
			}
			s := f.runtime(t)
			for _, target := range []struct {
				path  string
				agent bool
			}{{cliRuntimeThread, false}, {cliRuntimeAgents, true}} {
				if status, _ := cliRuntimeRequest(t, s, http.MethodGet, target.path, cliRuntimeTicket(target.agent), ""); status != wantStatus {
					t.Fatal("failed reader accepted credential", status)
				}
			}
			want := int32(1)
			if mode == "disabled" {
				want = 0
			}
			if f.legacy.Load() != 0 || f.thread.Load() != want || f.agents.Load() != want || f.other.Load() != 0 {
				t.Fatal("failed reader retried or fell back")
			}
		})
	}
}
