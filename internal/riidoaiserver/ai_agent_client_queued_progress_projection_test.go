package riidoaiserver

import (
	"context"
	"testing"
)

func TestClientVisibleProgressLinesHideBusyQueueStatus(t *testing.T) {
	t.Parallel()
	raw := []AgentThreadProgressLine{
		{Seq: 1, Message: "agent is busy; task assignment was queued"},
		{Seq: 2, Message: "Go 프로젝트 생성 수집 중"},
	}
	visible := copyClientVisibleProgressLines(raw)
	if len(visible) != 1 || visible[0].Seq != 2 {
		t.Fatalf("visible lines = %+v, want only runtime progress", visible)
	}
	if len(raw) != 2 || raw[0].Message != "agent is busy; task assignment was queued" {
		t.Fatalf("durable lines changed: %+v", raw)
	}
}

func TestListTaskThreadsHidesStaleBusyLineAfterRunning(t *testing.T) {
	t.Parallel()
	store := visibleDeletedAgentThreadStore()
	thread := &store.taskThreads["task-a"][0]
	thread.WorkStatus = AgentWorkStatusRunning
	thread.AssignmentState = AgentAssignmentStateRunning
	thread.Lines = []AgentThreadProgressLine{
		{Seq: 1, Message: clientMessageAgentBusyQueued},
		{Seq: 2, Message: "Go 프로젝트 생성 수집 중"},
	}

	response, err := store.ListAIAgentTaskThreads(context.Background(), visibleViewer(), "task-a")
	if err != nil {
		t.Fatal(err)
	}
	if got := response.Threads[0].Lines; len(got) != 1 || got[0].Seq != 2 {
		t.Fatalf("visible thread lines = %+v, want only running progress", got)
	}
	if got := store.taskThreads["task-a"][0].Lines; len(got) != 2 {
		t.Fatalf("durable thread lines changed: %+v", got)
	}
}

func TestLiveProgressFanoutHidesStaleBusyLine(t *testing.T) {
	t.Parallel()
	raw := AgentThreadProgressEvent{Lines: []AgentThreadProgressLine{
		{Seq: 1, Message: clientMessageAgentBusyQueued},
		{Seq: 2, Message: "Go 프로젝트 생성 수집 중"},
	}}
	visible, progress := clientEventForLiveFanout(ClientStreamEvent{Payload: raw})
	if !progress {
		t.Fatal("progress event must use progress fanout")
	}
	lines := visible.Payload.(AgentThreadProgressEvent).Lines
	if len(lines) != 1 || lines[0].Seq != 2 {
		t.Fatalf("visible SSE lines = %+v, want only running progress", lines)
	}
}
