package riidoaiserver

import "testing"

func TestWithoutSupersededQueuedClientEventsDropsStaleQueued(t *testing.T) {
	store := queuedFilterStore(
		queuedFilterThread("task", "thread-queued", "conversation-a", AgentWorkStatusQueued),
		queuedFilterThread("task", "thread-other", "conversation-b", AgentWorkStatusQueued),
		queuedFilterThread("task", "thread-running", "conversation-a", AgentWorkStatusRunning),
	)
	events := []ClientStreamEvent{
		queuedFilterStatusEvent(1, "task", "thread-queued", AgentTaskCommentQueuedByBusyAgent),
		queuedFilterStatusEvent(2, "task", "thread-other", AgentTaskCommentQueuedByBusyAgent),
		queuedFilterProgressEvent(3, "task", "thread-running"),
	}
	got := withoutSupersededQueuedClientEvents(store, events)
	if len(got) != 2 {
		t.Fatalf("expected stale queued event to be hidden, got %+v", got)
	}
	if got[0].Seq != 2 || got[1].Seq != 3 {
		t.Fatalf("unexpected retained event order: %+v", got)
	}
}

func TestQueuedClientEventIsSupersededByCurrentThreadState(t *testing.T) {
	store := queuedFilterStore(
		queuedFilterThread("task", "thread-running", "conversation-a", AgentWorkStatusRunning),
	)
	event := queuedFilterStatusEvent(1, "task", "thread-running", AgentTaskCommentQueuedByBusyAgent)
	if !queuedClientEventIsSupersededLocked(store, event, nil) {
		t.Fatalf("running thread must supersede stale queued event")
	}
}

func TestTaskThreadSupersedesQueuedEvent(t *testing.T) {
	cases := []struct {
		name   string
		thread AIAgentTaskThreadRecord
		want   bool
	}{
		{name: "queued", thread: AIAgentTaskThreadRecord{CommentKind: AgentTaskCommentQueuedByBusyAgent}, want: false},
		{name: "running_status", thread: AIAgentTaskThreadRecord{WorkStatus: AgentWorkStatusRunning}, want: true},
		{name: "running_state", thread: AIAgentTaskThreadRecord{AssignmentState: AgentAssignmentStateRunning}, want: true},
		{name: "progress_lines", thread: AIAgentTaskThreadRecord{Lines: []AgentThreadProgressLine{{Seq: 1}}}, want: true},
		{name: "agent_reply", thread: AIAgentTaskThreadRecord{CommentKind: AgentTaskCommentTaskCompleted}, want: true},
	}
	for _, tt := range cases {
		if got := taskThreadSupersedesQueuedEvent(tt.thread); got != tt.want {
			t.Fatalf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func queuedFilterStore(threads ...AIAgentTaskThreadRecord) *DevelopmentAIAgentClientStore {
	return &DevelopmentAIAgentClientStore{taskThreads: map[string][]AIAgentTaskThreadRecord{"task": threads}}
}

func queuedFilterThread(taskID, threadID, conversationID string, status AgentWorkStatus) AIAgentTaskThreadRecord {
	return AIAgentTaskThreadRecord{
		TaskID: taskID, ThreadID: threadID, ConversationID: conversationID,
		WorkStatus: status, AssignmentState: AgentAssignmentState(status),
	}
}

func queuedFilterStatusEvent(seq int64, taskID, threadID string, kind AgentTaskCommentKind) ClientStreamEvent {
	return ClientStreamEvent{Seq: seq, Payload: AgentWorkStatusChangedEvent{
		TaskID: taskID, ThreadID: threadID, CommentKind: kind,
	}}
}

func queuedFilterProgressEvent(seq int64, taskID, threadID string) ClientStreamEvent {
	return ClientStreamEvent{Seq: seq, Payload: AgentThreadProgressEvent{
		TaskID: taskID, ThreadID: threadID, Lines: []AgentThreadProgressLine{{Seq: 1}},
	}}
}
