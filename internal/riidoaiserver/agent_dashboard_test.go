package riidoaiserver

import (
	"fmt"
	"testing"
	"time"
)

func TestAgentDashboardHistoryUsesExecutionEventsAndKeepsMoreThanClientSnapshot(t *testing.T) {
	base := time.Date(2026, 10, 11, 14, 50, 0, 0, time.UTC)
	operations := []AssignmentOperationRecord{}
	for index := 0; index < 60; index++ {
		id := fmt.Sprintf("asn-%06d", index+1)
		for offset, state := range []AssignmentState{AssignmentQueued, AssignmentRunning, AssignmentCompleted} {
			record := replayOperationRecord("task-a", id, "agent-a", state, int64(index*3+offset+1), base.Add(time.Duration(offset)*10*time.Minute))
			record.Assignment.CreatedAt = base
			if state == AssignmentRunning {
				record.Events[0].Type = EventAssignmentRunning
			}
			operations = append(operations, record, record)
		}
	}
	state, err := stateFromAssignmentOperations(operations)
	if err != nil {
		t.Fatal(err)
	}
	history := dashboardHistoryFromSnapshot(snapshotFromState(&state, base), "agent-a")
	if len(history.Runs) != 60 {
		t.Fatalf("runs=%d", len(history.Runs))
	}
	view := buildAgentDashboard(AgentClientRecord{AgentID: "agent-a"}, history, map[string]AIAgentTaskContextComponent{"task-a": {ID: "task-a", Title: "Visible"}}, 2, base.Add(time.Hour))
	if view.CompletedRuns != 60 || view.Total != 60 || len(view.History) != 10 || view.Page != 2 || !view.HasNextPage {
		t.Fatalf("unexpected page: %+v", view)
	}
	if view.ActivityCompleted != 60 || view.ActiveDays != 1 {
		t.Fatalf("activity=%d days=%d", view.ActivityCompleted, view.ActiveDays)
	}
	if *view.History[0].DurationSeconds != 600 {
		t.Fatalf("queue included in duration: %+v", view.History[0])
	}
	last := view.Activity[len(view.Activity)-1]
	if last.Date != "2026-10-12" || last.Completed != 60 {
		t.Fatalf("Seoul boundary: %+v", last)
	}
	// A second reconstruction represents reopening the durable journal.
	reopened, err := stateFromAssignmentOperations(operations)
	if err != nil {
		t.Fatal(err)
	}
	if len(dashboardHistoryFromSnapshot(snapshotFromState(&reopened, base), "agent-a").Runs) != 60 {
		t.Fatal("history lost on reopen")
	}
}

func TestAgentDashboardFiltersBeforeCountsAndPaging(t *testing.T) {
	now := time.Now().UTC()
	history := AgentDashboardHistory{Runs: []AgentDashboardRun{
		{Assignment: Assignment{ID: "visible", TaskID: "allowed", State: AssignmentCompleted, CreatedAt: now}, EndedAt: &now},
		{Assignment: Assignment{ID: "secret", TaskID: "denied", State: AssignmentCompleted, CreatedAt: now}, EndedAt: &now},
		{Assignment: Assignment{ID: "queued-secret", TaskID: "denied", State: AssignmentQueued}, QueuePosition: 1},
		{Assignment: Assignment{ID: "queued-visible", TaskID: "allowed", State: AssignmentQueued}, QueuePosition: 2},
	}}
	view := buildAgentDashboard(AgentClientRecord{}, history, map[string]AIAgentTaskContextComponent{"allowed": {Title: "Visible"}}, 999, now)
	if view.Total != 1 || view.CompletedRuns != 1 || view.ActivityCompleted != 1 || view.WaitingCount != 1 || view.Page != 1 {
		t.Fatalf("inaccessible counts leaked: %+v", view)
	}
	if view.Queue[0].QueuePosition != 2 {
		t.Fatal("server queue order changed")
	}
	if view.History[0].DurationSeconds != nil {
		t.Fatal("unknown execution start fabricated")
	}
}
