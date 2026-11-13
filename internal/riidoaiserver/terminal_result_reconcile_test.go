package riidoaiserver

import "testing"

func TestHTTPCompletedResultSurvivesProjectionBeforeFanout(t *testing.T) {
	f, first, req, event := terminalResultFixture(t)
	assertTerminalResultHistory(t, readTerminalResultHistory(t, f))
	if err := f.client.RecordAIAgentAssignmentEvent(t.Context(), first.AgentID, req, event); err != nil {
		t.Fatal(err)
	}
	assertTerminalResultHistory(t, readTerminalResultHistory(t, f))
}

func seedMissingTerminalResult(t *testing.T, f *nativeResumeHTTPFixture, first AIAgentTaskActionResponse) {
	t.Helper()
	projection, found, err := f.store.LoadAssignmentProjection(t.Context(), first.AssignmentID)
	if err != nil || !found {
		t.Fatal("missing projection", err)
	}
	// Reproduce an already persisted row from the old state-only reconciler.
	projection = AssignmentProjection{Assignment: projection.Assignment, LastEventSeq: projection.LastEventSeq}
	if !f.client.applyAssignmentProjectionToTaskThread(AIAgentTaskThreadRecord{TaskID: first.TaskID, ThreadID: first.ThreadID, AssignmentID: first.AssignmentID}, projection, nil) {
		t.Fatal("failed to seed state-only completion")
	}
}

func TestHTTPCompletedResultRepairsHistoricalFallbackWithoutRerun(t *testing.T) {
	f, first, _, _ := terminalResultFixture(t)
	seedMissingTerminalResult(t, f, first)
	assertTerminalResultHistory(t, readTerminalResultHistory(t, f))
	assertTerminalResultHistory(t, readTerminalResultHistory(t, f))
	poll, err := f.store.PollAgent(t.Context(), first.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
	if err != nil || poll.Assignment != nil {
		t.Fatal("read repair must not start another execution", err)
	}
}

func TestCompletedResultLateFanoutRepairsFallback(t *testing.T) {
	f, first, req, event := terminalResultFixture(t)
	seedMissingTerminalResult(t, f, first)
	for range 2 {
		if err := f.client.RecordAIAgentAssignmentEvent(t.Context(), first.AgentID, req, event); err != nil {
			t.Fatal(err)
		}
	}
	threads, err := f.client.ListAIAgentTaskThreads(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, first.TaskID)
	if err != nil || len(threads.Threads) != 1 || threads.Threads[0].ResultMessage != terminalResultBody {
		t.Fatalf("late completion left a fallback: %+v, %v", threads, err)
	}
}
