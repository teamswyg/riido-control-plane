package riidoaiserver

import (
	"testing"
	"time"
)

func TestAgentDashboardPreservesRequesterPerExecution(t *testing.T) {
	now := time.Now().UTC()
	history := AgentDashboardHistory{Runs: []AgentDashboardRun{
		{Assignment: Assignment{ID: "requested", TaskID: "allowed", State: AssignmentCompleted, CreatedBy: "requester", CreatedAt: now}},
		{Assignment: Assignment{ID: "legacy", TaskID: "allowed", State: AssignmentCompleted, CreatedAt: now}},
		{Assignment: Assignment{ID: "private", TaskID: "denied", State: AssignmentCompleted, CreatedBy: "private-user", CreatedAt: now}},
	}}
	view := buildAgentDashboard(AgentClientRecord{AgentID: "agent-a"}, history, map[string]AIAgentTaskContextComponent{"allowed": {Title: "Visible"}}, 1, now)
	if len(view.History) != 2 {
		t.Fatalf("unexpected records: %+v", view.History)
	}
	for _, row := range view.History {
		want := ""
		if row.AssignmentID == "requested" {
			want = "requester"
		}
		if row.CreatedBy != want {
			t.Fatalf("requester changed or fabricated: %+v", row)
		}
	}
}
