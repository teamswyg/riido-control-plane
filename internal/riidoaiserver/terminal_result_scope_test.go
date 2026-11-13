package riidoaiserver

import "testing"

func TestCompletedResultRepairIsTaskScopedAndPreservesNeedsInput(t *testing.T) {
	f, first, req, event := terminalResultFixture(t)
	seedMissingTerminalResult(t, f, first)
	reader := &recordingAssignmentProjectionReader{}
	for _, taskID := range []string{"", "another-task"} {
		_, err := f.client.ReconcileAIAgentActiveThreadProjections(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, taskID, reader)
		if err != nil || len(reader.loadedAssignmentIDs()) != 0 {
			t.Fatal("unscoped read scanned historical terminal results", err)
		}
	}
	event.Metadata = map[string]string{"assignment_result_status": "needs_input"}
	event.Message = "어느 저장소를 사용할까요?"
	if err := f.client.RecordAIAgentAssignmentEvent(t.Context(), first.AgentID, req, event); err != nil {
		t.Fatal(err)
	}
	thread := f.client.taskThreads[first.TaskID][0]
	if thread.AssignmentState != AgentAssignmentStateWaiting || thread.WorkStatus != AgentWorkStatusWaitingForUser ||
		thread.CommentKind != AgentTaskCommentNeedsInput || thread.ResultMessage != event.Message || taskThreadHasActiveStream(thread) {
		t.Fatalf("recovered result lost user action metadata or reopened runtime: %+v", thread)
	}
}
