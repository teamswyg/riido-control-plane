package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDashboardSubscriberReceivesCommittedRemoteQueueEventOnce(t *testing.T) {
	const agent = "agent-owned-codex"
	source, replica := NewDevelopmentAIAgentClientStore(), NewDevelopmentAIAgentClientStore()
	p := AuthorizationResult{WorkspaceID: defaultAIAgentClientWorkspaceID, PrincipalID: "user-1"}
	ctx := context.WithValue(context.Background(), dashboardStreamScope{}, agent)
	_, events, stop, err := replica.SubscribeAIAgentClientEvents(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	source.mu.Lock()
	source.appendClientEventLocked(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{EventType: AgentClientEventWorkStatusChanged, TaskID: "t", AgentID: agent, AssignmentState: AgentAssignmentStateQueued})
	source.mu.Unlock()
	snap, err := source.snapshot(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = replica.restoreSnapshotPreservingSubscribers(snap); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case e := <-events:
		if event, ok := e.Payload.(AgentWorkStatusChangedEvent); !ok || event.AgentID != agent || event.AssignmentState != AgentAssignmentStateQueued {
			t.Fatal("wrong restored event")
		}
	default:
		t.Fatal("committed remote event was not delivered")
	}
	select {
	case <-events:
		t.Fatal("reloading the same snapshot duplicated a transition")
	default:
	}
}
