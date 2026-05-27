package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func TestHTTPProviderStatusSyncAndRead(t *testing.T) {
	syncedAt := time.Date(2026, 5, 27, 15, 30, 0, 0, time.UTC)
	store := newProviderStatusTestStore(syncedAt)
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "daemon:agent-a",
		Token:       "agent-token",
		Scopes: []string{
			"agent:agent-a:provider-status:write",
			"agent:agent-a:provider-status:read",
		},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	server := NewServer(ServerConfig{ProviderStatus: store, Authorizer: authorizer}).Handler()

	body := `{"daemon_id":" daemon-a ","device_id":" device-a ","runtime_id":" runtime-a ","distribution_channel":"msix-store","app_version":" 1.2.3 ","providers":[{"provider_kind":"cursor","routing_status":"login-required"},{"provider_kind":"codex","routing_status":"available"}]}`
	postReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/provider-status", strings.NewReader(body))
	postReq.Header.Set("Authorization", "Bearer agent-token")
	postResp := httptest.NewRecorder()
	server.ServeHTTP(postResp, postReq)
	if postResp.Code != http.StatusOK {
		t.Fatalf("provider status sync status=%d body=%s", postResp.Code, postResp.Body.String())
	}
	var synced ProviderStatusSyncResponse
	if err := json.Unmarshal(postResp.Body.Bytes(), &synced); err != nil {
		t.Fatalf("sync json: %v", err)
	}
	if synced.SchemaVersion != SchemaVersion || synced.AgentID != "agent-a" || synced.DaemonID != "daemon-a" || synced.RuntimeID != "runtime-a" || synced.AppVersion != "1.2.3" {
		t.Fatalf("synced response = %+v", synced)
	}
	if synced.DistributionChannel != hostintegration.DistributionChannelMSIXStore {
		t.Fatalf("distribution channel = %q", synced.DistributionChannel)
	}
	if got, want := providerStatusKinds(synced.Providers), []string{"codex", "cursor"}; !sameStrings(got, want) {
		t.Fatalf("synced providers = %v, want %v", got, want)
	}
	if !synced.SyncedAt.Equal(syncedAt) {
		t.Fatalf("synced_at = %s, want %s", synced.SyncedAt, syncedAt)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/agents/agent-a/provider-status", nil)
	getReq.Header.Set("Authorization", "Bearer agent-token")
	getResp := httptest.NewRecorder()
	server.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusOK {
		t.Fatalf("provider status read status=%d body=%s", getResp.Code, getResp.Body.String())
	}
	var read ProviderStatusSyncResponse
	if err := json.Unmarshal(getResp.Body.Bytes(), &read); err != nil {
		t.Fatalf("read json: %v", err)
	}
	if read.AgentID != synced.AgentID || read.DaemonID != synced.DaemonID || !read.SyncedAt.Equal(syncedAt) {
		t.Fatalf("read response = %+v, want sync %+v", read, synced)
	}
}

func TestHTTPProviderStatusRejectsPrivateAndUnknownFields(t *testing.T) {
	store := newProviderStatusTestStore(time.Date(2026, 5, 27, 15, 30, 0, 0, time.UTC))
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "daemon:agent-a",
		Token:       "agent-token",
		Scopes:      []string{"agent:agent-a:provider-status:write"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	server := NewServer(ServerConfig{ProviderStatus: store, Authorizer: authorizer}).Handler()

	body := `{"daemon_id":"daemon-a","runtime_id":"runtime-a","distribution_channel":"developer-id","providers":[{"provider_kind":"codex","routing_status":"available","executable_path":"/private/bin/codex"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/provider-status", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer agent-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("private field status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHTTPProviderStatusRequiresScopedAuthorization(t *testing.T) {
	store := newProviderStatusTestStore(time.Date(2026, 5, 27, 15, 30, 0, 0, time.UTC))
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "daemon:agent-b",
		Token:       "agent-token",
		Scopes:      []string{"agent:agent-b:provider-status:write"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	server := NewServer(ServerConfig{ProviderStatus: store, Authorizer: authorizer}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/provider-status", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer agent-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("other agent sync status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func providerStatusKinds(providers []ProviderStatusRecord) []string {
	out := make([]string, 0, len(providers))
	for _, provider := range providers {
		out = append(out, string(provider.ProviderKind))
	}
	return out
}
