package riidoaiserver

import (
	"context"
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"path/filepath"
	"testing"
	"time"
)

func TestAgentDashboardLifecycleAndDurableRestart(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	snapshot, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "assignments.json"))
	if err != nil {
		t.Fatal(err)
	}
	config := StoreConfig{SnapshotStore: snapshot, Now: func() time.Time { return now }}
	store, err := OpenStoreWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	assignment, err := store.AssignTask(ctx, "task", AssignRequest{ComponentID: "task", AgentID: "agent", RuntimeProvider: "codex", Prompt: "work"})
	if err != nil {
		t.Fatal(err)
	}
	load := func() AgentDashboardResponse {
		t.Helper()
		history, err := store.LoadAgentDashboardHistory(ctx, "agent")
		if err != nil {
			t.Fatal(err)
		}
		return buildAgentDashboard(AgentClientRecord{AgentID: "agent"}, history, map[string]AIAgentTaskContextComponent{"task": {ID: "task", Title: "Task"}}, 1, now)
	}
	if view := load(); view.WaitingCount != 1 || view.Current != nil {
		t.Fatalf("queued=%+v", view)
	}
	now = now.Add(5 * time.Minute)
	for _, state := range []AssignmentState{AssignmentLeased, AssignmentReady, AssignmentRunning} {
		_, err = store.RecordAgentEvent(ctx, "agent", AgentEventRequest{AssignmentID: assignment.ID, TaskID: "task", State: state})
		if err != nil {
			t.Fatal(err)
		}
	}
	now = now.Add(2 * time.Minute)
	_, err = store.RecordAgentEvent(ctx, "agent", AgentEventRequest{AssignmentID: assignment.ID, TaskID: "task", EventType: assignmentcontract.EventAssignmentWaitingForUser, ConfirmationReason: assignmentcontract.ConfirmationApprovalRequired, Message: "Approve execution"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.RecordAgentEvent(ctx, "agent", AgentEventRequest{AssignmentID: assignment.ID, TaskID: "task", EventType: EventRiidoLog, Message: "background log"})
	if err != nil {
		t.Fatal(err)
	}
	if view := load(); view.Current.Message != "Approve execution" {
		t.Fatalf("confirmation overwritten: %+v", view.Current)
	}
	if view := load(); view.ConfirmationCount != 1 || view.Current.ConfirmationReason != "approval_required" {
		t.Fatalf("waiting=%+v", view.Current)
	}
	now = now.Add(3 * time.Minute)
	_, err = store.RecordAgentEvent(ctx, "agent", AgentEventRequest{AssignmentID: assignment.ID, TaskID: "task", EventType: EventAssignmentRunning})
	if err != nil {
		t.Fatal(err)
	}
	if view := load(); view.ConfirmationCount != 0 || view.Current.State != "running" {
		t.Fatalf("resumed=%+v", view.Current)
	}
	now = now.Add(5 * time.Minute)
	for index := 0; index < 2; index++ {
		_, err = store.RecordAgentEvent(ctx, "agent", AgentEventRequest{AssignmentID: assignment.ID, TaskID: "task", State: AssignmentCompleted})
		if err != nil {
			t.Fatal(err)
		}
	}
	view := load()
	if view.CompletedRuns != 1 || view.Current != nil || view.History[0].DurationSeconds == nil || *view.History[0].DurationSeconds != 600 {
		t.Fatalf("completed=%+v history=%+v", view, view.History)
	}
	store.Close()
	store, err = OpenStoreWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	view = load()
	if view.CompletedRuns != 1 || !view.HistoryDurable || *view.History[0].DurationSeconds != 600 {
		t.Fatalf("reopened=%+v", view)
	}
}
