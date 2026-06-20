package riidoaiserver

import (
	"context"
	"testing"
)

func countWorkStatusChangedEventsForTask(t *testing.T, store *DevelopmentAIAgentClientStore, principal AuthorizationResult, taskID string) int {
	t.Helper()
	events, err := store.AIAgentClientEvents(context.Background(), principal)
	if err != nil {
		t.Fatalf("AIAgentClientEvents: %v", err)
	}
	count := 0
	for _, event := range events {
		if event.EventType != AgentClientEventWorkStatusChanged {
			continue
		}
		status, ok := event.Payload.(AgentWorkStatusChangedEvent)
		if ok && status.TaskID == taskID {
			count++
		}
	}
	return count
}

func lastWorkStatusChangedEventForTask(t *testing.T, store *DevelopmentAIAgentClientStore, principal AuthorizationResult, taskID string) (AgentWorkStatusChangedEvent, bool) {
	t.Helper()
	events, err := store.AIAgentClientEvents(context.Background(), principal)
	if err != nil {
		t.Fatalf("AIAgentClientEvents: %v", err)
	}
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].EventType != AgentClientEventWorkStatusChanged {
			continue
		}
		status, ok := events[i].Payload.(AgentWorkStatusChangedEvent)
		if ok && status.TaskID == taskID {
			return status, true
		}
	}
	return AgentWorkStatusChangedEvent{}, false
}
