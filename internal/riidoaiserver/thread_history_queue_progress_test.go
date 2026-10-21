package riidoaiserver

import "testing"

func TestV3RunningHistoryHidesPersistedQueueStatusAndKeepsProgress(t *testing.T) {
	store := visibleDeletedAgentThreadStore()
	thread := &store.taskThreads["task-a"][0]
	thread.WorkStatus = AgentWorkStatusRunning
	thread.AssignmentState = AgentAssignmentStateRunning
	thread.CommentKind = AgentTaskCommentRuntimeProgress
	thread.Message = clientMessageAgentBusyQueued
	thread.Lines = []AgentThreadProgressLine{
		{Seq: 1, Message: clientMessageAgentBusyQueued},
		{Seq: 2, Message: "agent assignment is queued"},
		{Seq: 3, Message: "Inspecting the requested files"},
	}
	response, err := store.ListAIAgentTaskThreadHistory(t.Context(), visibleViewer(), "task-a")
	if err != nil || len(response.Threads) != 1 {
		t.Fatalf("history: %v", err)
	}
	foundProgress := false
	for _, message := range response.Threads[0].Messages {
		if message.Body == clientMessageAgentBusyQueued || message.Body == clientMessageTaskQueued || message.Body == "agent assignment is queued" {
			t.Fatalf("v3 history leaked queue status: %+v", message)
		}
		if message.Body == "Inspecting the requested files" {
			foundProgress = true
		}
	}
	if !foundProgress {
		t.Fatal("v3 history lost provider progress")
	}
	if len(store.taskThreads["task-a"][0].Lines) != 3 {
		t.Fatal("read changed persisted lines")
	}
}
