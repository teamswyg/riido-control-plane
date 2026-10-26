package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAgentDirectoryAdminRequesterSeesEveryWorkspaceAgent(t *testing.T) {
	contexts := NewMemoryAssignmentContextStore()
	f := newAssignmentContextFixture(t, contexts)
	snapshot := contexts.snapshots[f.leased.ID]
	if snapshot.RequesterAdmin {
		t.Fatal("a non-admin requester must not be recorded as admin")
	}
	snapshot.RequesterAdmin = true
	contexts.snapshots[f.leased.ID] = snapshot
	code, body := f.read(AssignmentContextViewAgents, "", f.leased.LeaseToken)
	var resp struct{ Data AgentDirectoryPage }
	if code != http.StatusOK || json.Unmarshal([]byte(body), &resp) != nil || resp.Data.Total != 4 || !strings.Contains(body, "agent-private-cursor") {
		t.Fatalf("admin directory status=%d body=%s", code, body)
	}
	if strings.Contains(body, "instruction") || strings.Contains(body, "owner_principal_id") {
		t.Fatalf("admin directory leaked private fields: %s", body)
	}
	if code, _ := f.read(AssignmentContextViewAgent, "agent-private-cursor", f.leased.LeaseToken); code != http.StatusOK {
		t.Fatalf("admin show status=%d", code)
	}
	// Without a recorded admin fact (no snapshot), the directory stays narrow.
	delete(contexts.snapshots, f.leased.ID)
	if code, body := f.read(AssignmentContextViewAgents, "", f.leased.LeaseToken); code != http.StatusOK || strings.Contains(body, "agent-private-cursor") {
		t.Fatalf("missing snapshot widened scope: status=%d body=%s", code, body)
	}
}

func TestCaptureRecordsRequesterAdminAtRequestTime(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryAssignmentContextStore()
	server := NewServer(ServerConfig{AssignmentContext: store})
	assignment := Assignment{ID: "asg-admin", TaskID: "task-1", AgentID: "agent-a", CreatedBy: "admin-1"}
	server.captureAssignmentContext(ctx, assignment, &AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task-1"}}, assignmentContextTrigger{Kind: AssignmentContextTriggerAssign, ID: "req", RequesterAdmin: true})
	got, found, err := store.LoadAssignmentContextSnapshot(ctx, "asg-admin")
	if err != nil || !found || !got.RequesterAdmin {
		t.Fatalf("snapshot=%+v found=%v err=%v", got, found, err)
	}
}
