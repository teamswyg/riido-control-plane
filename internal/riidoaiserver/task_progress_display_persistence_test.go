package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestContextualProgressSurvivesStoredThreadAndEventJSON(t *testing.T) {
	line := AgentThreadProgressLine{Seq: 1, Message: "rendered progress", MessageCode: 1103, MessageArgs: map[string]string{"description": "답글 전송 권한을 점검하고 있어요."}}
	thread := AIAgentTaskThreadRecord{ThreadID: "thread", Lines: []AgentThreadProgressLine{line}}
	event := AgentThreadProgressEvent{ThreadID: "thread", Lines: []AgentThreadProgressLine{line}}
	threadJSON, _ := json.Marshal(thread)
	eventJSON, _ := json.Marshal(event)
	var restoredThread AIAgentTaskThreadRecord
	var restoredEvent AgentThreadProgressEvent
	if err := json.Unmarshal(threadJSON, &restoredThread); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(eventJSON, &restoredEvent); err != nil {
		t.Fatal(err)
	}
	for _, restored := range []AgentThreadProgressLine{restoredThread.Lines[0], restoredEvent.Lines[0]} {
		if got := taskProgressDisplayMessage(restored); got != taskProgressDisplayMessage(line) {
			t.Fatalf("stored JSON lost display summary: %q", got)
		}
	}
	rows := taskThreadProgressMessages(restoredThread)
	if rows[0].DisplayMessage != taskProgressDisplayMessage(line) {
		t.Fatal("cold history lost restored display summary")
	}
}

func TestContextualProgressRejectsClientSuppliedDisplayMessage(t *testing.T) {
	var batch AgentThreadProgressBatchRequest
	if err := json.Unmarshal([]byte(`{"lines":[{"seq":1,"message":"unknown raw output","display_message":"invented success"}]}`), &batch); err != nil {
		t.Fatal(err)
	}
	if got := taskProgressDisplayMessage(batch.Lines[0]); got != "" {
		t.Fatalf("external ingest trusted display_message: %q", got)
	}
}

func TestContextualProgressCacheInvalidatesChangedSummary(t *testing.T) {
	line := AgentThreadProgressLine{Seq: 1, Message: "same rendered body", MessageCode: 1103, MessageArgs: map[string]string{"description": "댓글 구조를 살펴보고 있어요."}}
	thread := AIAgentTaskThreadRecord{ThreadID: "thread", Lines: []AgentThreadProgressLine{line}}
	cache := newTaskThreadProgressMessageCache(thread, taskThreadProgressMessages(thread))
	thread.Lines[0].MessageArgs = map[string]string{"description": "답글 전송 권한을 점검하고 있어요."}
	if cache.matches(thread) {
		t.Fatal("cache reused stale contextual summary")
	}
}
