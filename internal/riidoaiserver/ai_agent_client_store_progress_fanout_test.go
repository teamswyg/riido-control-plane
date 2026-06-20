package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDevelopmentAIAgentClientStoreThreadProgressFanout(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	history, events, cancel, err := store.SubscribeAIAgentClientEvents(context.Background(), AuthorizationResult{PrincipalID: "user-1"})
	if err != nil {
		t.Fatalf("SubscribeAIAgentClientEvents: %v", err)
	}
	defer cancel()
	if len(history) == 0 {
		t.Fatal("expected replay history")
	}
	if _, err := store.RecordAIAgentThreadProgress(context.Background(), "agent-owned-codex", AgentThreadProgressBatchRequest{
		AssignmentID: "asn-1",
		TaskID:       "task-1",
		RunID:        "run-1",
		Lines:        []AgentThreadProgressLine{{Seq: 1, Message: "검증 실행 중 - /tmp/riido-runtime/log.txt"}},
	}); err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
	}
	select {
	case event := <-events:
		progress, ok := event.Payload.(AgentThreadProgressEvent)
		if !ok || progress.EventType != AgentClientEventThreadProgress || progress.Lines[0].Message != "검증 실행 중 - 로컬 파일" {
			t.Fatalf("fanout event = %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for progress fanout")
	}
}
