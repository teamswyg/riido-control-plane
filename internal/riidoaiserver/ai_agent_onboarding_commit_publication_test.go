package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type onboardingFailingSnapshot struct {
	memoryAIAgentClientSnapshotStore
	fail bool
}

func (s *onboardingFailingSnapshot) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if s.fail {
		return errors.New("onboarding fixture persistence unavailable")
	}
	return s.memoryAIAgentClientSnapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot)
}

func TestOnboardingCompletionDoesNotPublishBeforeDurableCommit(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	snapshots := &onboardingFailingSnapshot{}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("publication failure"))
	if err != nil {
		t.Fatal(err)
	}
	_, live, unsubscribe, err := store.SubscribeAIAgentClientEvents(ctx, principal)
	if err != nil {
		t.Fatal(err)
	}
	defer unsubscribe()
	snapshots.fail = true
	request := CompleteAIAgentOnboardingTaskRequest{AgentID: created.Agent.AgentID, TemplateKey: AIAgentOnboardingTemplateSchedule, RequestID: "durable-completion"}
	_, completionErr := store.CompleteAIAgentOnboardingTask(ctx, principal, "task-not-yet-committed", request)
	if completionErr == nil {
		t.Fatal("completion acknowledged a failed durable commit")
	}
	rec := httptest.NewRecorder()
	writeAIAgentClientError(rec, completionErr)
	if rec.Code != http.StatusServiceUnavailable || strings.Contains(rec.Body.String(), "fixture persistence") {
		t.Fatal("completion dependency failure lost retry status or exposed raw detail")
	}
	select {
	case <-live:
		t.Fatal("uncommitted onboarding completion reached a live subscriber")
	default:
	}
	if store.hasOnboardingCompletionRecord("task-not-yet-committed", created.Agent.AgentID) {
		t.Fatal("failed completion retained local one-time state")
	}
	snapshots.fail = false
	if _, err := store.CompleteAIAgentOnboardingTask(ctx, principal, "task-not-yet-committed", request); err != nil {
		t.Fatal(err)
	}
	select {
	case <-live:
	default:
		t.Fatal("committed completion was not published")
	}
}
