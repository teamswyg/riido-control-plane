package riidoaiserver

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDashboardStreamQueuesAreScopedAndAuthorized(t *testing.T) {
	base := NewDevelopmentAIAgentClientStore()
	server := httptest.NewServer(NewServer(ServerConfig{AIAgentClient: base, Authorizer: dashboardTestAuthorizer{}, TaskContext: &dashboardTestContext{}}).Handler())
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/v2/client/workspaces/"+defaultAIAgentClientWorkspaceID+"/ai-agent/events?dashboard_agent_id=agent-owned-codex", nil)
	req.Header.Set("Authorization", "Bearer viewer-token")
	res, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	reader := bufio.NewScanner(res.Body)
	next := func() map[string]any {
		t.Helper()
		for reader.Scan() {
			if raw, ok := strings.CutPrefix(reader.Text(), "data: "); ok {
				var v map[string]any
				if err := json.Unmarshal([]byte(raw), &v); err != nil {
					t.Fatal(err)
				}
				return v
			}
		}
		t.Fatal("stream closed", reader.Err())
		return nil
	}
	if !reader.Scan() || reader.Text() != ": dashboard stream ready" {
		t.Fatal("live subscription not ready")
	}

	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	_, ordinary, stop, err := base.SubscribeAIAgentClientEvents(ctx, principal)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	send := func(taskID, id string) {
		base.mu.Lock()
		defer base.mu.Unlock()
		base.appendClientEventLocked(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{EventType: AgentClientEventWorkStatusChanged, AgentID: id, TaskID: taskID, AssignmentID: "asn", AssignmentState: AgentAssignmentStateQueued})
	}
	send("secret", "agent-owned-codex")
	send("other", "other-agent")
	send("visible", "agent-owned-codex")
	if q := next(); q["task_id"] != "visible" || q["assignment_state"] != "queued" {
		t.Fatalf("unexpected event: %v", q)
	}
	select {
	case <-ordinary:
		t.Fatal("legacy stream received dashboard queue")
	default:
	}
}
