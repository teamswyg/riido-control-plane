package riidoaiserver

import (
	"testing"
	"time"
)

func TestFallbackAgentThreadProgressResponseUsesProvidedThreadID(t *testing.T) {
	req := fallbackProgressRequest("thread-explicit")
	got := fallbackAgentThreadProgressResponse("agent-a", req)

	if got.AcceptedLines != len(req.Lines) {
		t.Fatalf("accepted lines = %d, want %d", got.AcceptedLines, len(req.Lines))
	}
	if got.Event.ThreadID != "thread-explicit" {
		t.Fatalf("thread id = %q, want explicit thread id", got.Event.ThreadID)
	}
	assertFallbackProgressEvent(t, got.Event, req)
}

func TestFallbackAgentThreadProgressResponseDerivesThreadID(t *testing.T) {
	req := fallbackProgressRequest("")
	got := fallbackAgentThreadProgressResponse("agent-a", req)
	wantThread := threadIDForRun(req.TaskID, "agent-a", req.RunID)

	if got.Event.ThreadID != wantThread {
		t.Fatalf("thread id = %q, want %q", got.Event.ThreadID, wantThread)
	}
	assertFallbackProgressEvent(t, got.Event, req)
}

func fallbackProgressRequest(threadID string) AgentThreadProgressBatchRequest {
	started := time.Date(2026, 7, 7, 1, 2, 3, 0, time.UTC)
	return AgentThreadProgressBatchRequest{
		TaskID:         "task-a",
		ThreadID:       threadID,
		RunID:          "run-a",
		BatchStartedAt: started,
		BatchEndedAt:   started.Add(time.Second),
		Lines:          []AgentThreadProgressLine{{Seq: 1, Message: "working"}},
	}
}

func assertFallbackProgressEvent(t *testing.T, got AgentThreadProgressEvent, req AgentThreadProgressBatchRequest) {
	t.Helper()
	if got.EventType != AgentClientEventThreadProgress ||
		got.SchemaVersion != SchemaVersion ||
		got.TaskID != req.TaskID ||
		got.RunID != req.RunID ||
		got.WorkStatus != AgentWorkStatusRunning ||
		got.AssignmentState != AgentAssignmentStateRunning ||
		got.CommentKind != AgentTaskCommentRuntimeProgress {
		t.Fatalf("unexpected fallback event: %+v", got)
	}
	if len(got.Lines) != 1 || got.Lines[0].Message != "working" {
		t.Fatalf("unexpected lines: %+v", got.Lines)
	}
}
