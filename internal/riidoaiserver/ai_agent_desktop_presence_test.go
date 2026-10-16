package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPDesktopPresenceReportsOnlyTheSameAccount(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{
		{PrincipalID: "web-user", Token: "web-token", Scopes: []string{"ai-agent:*"}},
		{PrincipalID: "other-user", Token: "other-token", Scopes: []string{"ai-agent:*"}},
	})
	if body := readDesktopPresence(t, server, "web-token"); body.LastSeenAt != nil || body.CheckedAt.IsZero() {
		t.Fatalf("presence before any desktop connection = %+v", body)
	}

	connectDesktopForPresence(t, server, "other-token", "machine-other")
	if body := readDesktopPresence(t, server, "web-token"); body.LastSeenAt != nil {
		t.Fatalf("another account's desktop counted as presence: %+v", body)
	}

	connectedAfter := time.Now().UTC()
	connectDesktopForPresence(t, server, "web-token", "machine-web")
	body := readDesktopPresence(t, server, "web-token")
	if body.LastSeenAt == nil || body.LastSeenAt.Before(connectedAfter) || body.LastSeenAt.After(body.CheckedAt) {
		t.Fatalf("presence after desktop connection = %+v (connected after %s)", body, connectedAfter)
	}
}

func TestHTTPDesktopPresenceRejectsWrites(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "web-user", Token: "web-token", Scopes: []string{"ai-agent:*"},
	}})
	req := httptest.NewRequest(http.MethodPost, desktopPresenceTestPath, nil)
	req.Header.Set(aiAgentTokenHeader, "web-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST desktop presence status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestPersistentDesktopPresenceSeesConnectionsFromAnotherInstance(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	reader, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open reader: %v", err)
	}
	writer, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	principal := AuthorizationResult{PrincipalID: "web-user", WorkspaceID: "workspace-beta"}
	if lastSeenAt, err := reader.DesktopPresenceLastSeenAt(ctx, principal); err != nil || !lastSeenAt.IsZero() {
		t.Fatalf("presence before connect = %s, err = %v", lastSeenAt, err)
	}
	if _, err := writer.ConnectAIAgentDevice(ctx, principal, "machine-web"); err != nil {
		t.Fatalf("connect: %v", err)
	}
	reloadedAt := reader.snapshotLastReloadAt
	lastSeenAt, err := reader.DesktopPresenceLastSeenAt(ctx, principal)
	if err != nil || lastSeenAt.IsZero() {
		t.Fatalf("reader presence right after another instance connected = %s, err = %v", lastSeenAt, err)
	}
	// Restoring the snapshot would drop in-memory daemon heartbeats saved only every few minutes.
	if !reader.snapshotLastReloadAt.Equal(reloadedAt) {
		t.Fatalf("presence read restored the snapshot into memory")
	}
}
