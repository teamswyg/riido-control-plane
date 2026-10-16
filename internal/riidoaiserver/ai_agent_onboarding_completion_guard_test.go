package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func onboardingGuardPrincipal() AuthorizationResult {
	return AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
}

func onboardingGuardCreate(t *testing.T, store *DevelopmentAIAgentClientStore, principal AuthorizationResult, name string, visibility AgentVisibility) string {
	t.Helper()
	req := profileColorTestRequest(name)
	req.Visibility = visibility
	created, err := store.CreateAIAgentFromOnboardingFixture(context.Background(), principal, "riido_pm", req)
	if err != nil {
		t.Fatalf("CreateAIAgentFromOnboardingFixture(%s): %v", name, err)
	}
	return created.Agent.AgentID
}

func onboardingGuardComplete(store *DevelopmentAIAgentClientStore, principal AuthorizationResult, taskID, agentID, requestID string) (AIAgentTaskActionResponse, error) {
	return store.CompleteAIAgentOnboardingTask(context.Background(), principal, taskID, CompleteAIAgentOnboardingTaskRequest{
		AgentID:     agentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
		RequestID:   requestID,
	})
}

func TestOnboardingCompletionRequiresOnboardingProvenance(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	_, err := onboardingGuardComplete(store, onboardingGuardPrincipal(), "task-guard-provenance", "agent-owned-codex", "")
	if !errors.Is(err, errOnboardingCompletionIneligible) {
		t.Fatalf("seeded agent completion error = %v, want ineligible", err)
	}
}

func TestOnboardingCompletionRejectsAgentOwnedByAnotherPrincipal(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	agentID := onboardingGuardCreate(t, store, onboardingGuardPrincipal(), "public onboarding", AgentVisibilityPublic)
	other := AuthorizationResult{PrincipalID: "user-2", WorkspaceID: defaultAIAgentClientWorkspaceID}
	_, err := onboardingGuardComplete(store, other, "task-guard-owner", agentID, "")
	if !errors.Is(err, ErrAIAgentNotFound) {
		t.Fatalf("non-owner completion error = %v, want not found", err)
	}
}

func TestOnboardingCompletionIsOneTimePerAgent(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := onboardingGuardPrincipal()
	agentID := onboardingGuardCreate(t, store, principal, "one time", AgentVisibilityPrivate)
	if _, err := onboardingGuardComplete(store, principal, "task-once-1", agentID, ""); err != nil {
		t.Fatalf("first completion: %v", err)
	}
	_, err := onboardingGuardComplete(store, principal, "task-once-2", agentID, "")
	if !errors.Is(err, errOnboardingCompletionConflict) {
		t.Fatalf("second task completion error = %v, want conflict", err)
	}
}

func TestOnboardingCompletionRejectsTaskWithPriorAssignmentHistory(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := onboardingGuardPrincipal()
	agentID := onboardingGuardCreate(t, store, principal, "history", AgentVisibilityPrivate)
	if _, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-guard-history", AssignAIAgentTaskRequest{AgentID: agentID, AssignmentID: "asn-prior"}); err != nil {
		t.Fatalf("prior assignment: %v", err)
	}
	_, err := onboardingGuardComplete(store, principal, "task-guard-history", agentID, "")
	if !errors.Is(err, errOnboardingCompletionConflict) {
		t.Fatalf("completion after prior assignment error = %v, want conflict", err)
	}
	if _, err := onboardingGuardComplete(store, principal, "task-guard-fresh", agentID, ""); errors.Is(err, errOnboardingCompletionConflict) {
		t.Fatalf("a rejected prior-history attempt must not consume the one-time completion: %v", err)
	}
}

func TestOnboardingCompletionRejectsActiveOrQueuedExecutionOnSameTask(t *testing.T) {
	for _, state := range []AssignmentState{AssignmentQueued, AssignmentRunning} {
		t.Run(string(state), func(t *testing.T) {
			ctx := context.Background()
			store := NewDevelopmentAIAgentClientStore()
			principal := onboardingGuardPrincipal()
			agentID := onboardingGuardCreate(t, store, principal, "active", AgentVisibilityPrivate)
			if _, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-guard-active", AssignAIAgentTaskRequest{
				AgentID:      agentID,
				AssignmentID: "asn-active",
				durableState: state,
			}); err != nil {
				t.Fatalf("active assignment: %v", err)
			}
			_, err := onboardingGuardComplete(store, principal, "task-guard-active", agentID, "")
			if !errors.Is(err, errOnboardingCompletionConflict) {
				t.Fatalf("completion during %s execution error = %v, want conflict", state, err)
			}
		})
	}
}

func TestOnboardingCompletionRequestIDRejectsReuseForAnotherTarget(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := onboardingGuardPrincipal()
	first := onboardingGuardCreate(t, store, principal, "request first", AgentVisibilityPrivate)
	second := onboardingGuardCreate(t, store, principal, "request second", AgentVisibilityPrivate)
	original, err := onboardingGuardComplete(store, principal, "task-req-1", first, "req-shared")
	if err != nil {
		t.Fatalf("first completion: %v", err)
	}
	replay, err := onboardingGuardComplete(store, principal, "task-req-1", first, "req-shared")
	if err != nil || replay.ThreadID != original.ThreadID {
		t.Fatalf("same-target replay = %+v, %v; want original record", replay, err)
	}
	_, err = onboardingGuardComplete(store, principal, "task-req-2", second, "req-shared")
	if !errors.Is(err, errOnboardingCompletionConflict) || !strings.Contains(err.Error(), "request_id") {
		t.Fatalf("request_id reuse for another target error = %v, want request_id conflict", err)
	}
	store.mu.Lock()
	_, recorded := store.taskThreadForAssignmentLocked("task-req-2", second, onboardingSeedAssignmentID("task-req-2", second))
	store.mu.Unlock()
	if recorded {
		t.Fatal("rejected request_id reuse must not record a completion on the other task")
	}
}

// CP1: onboarding and assignment receipts share no namespace in either order.
func TestOnboardingCompletionReceiptNamespaceIsSeparateFromAssignment(t *testing.T) {
	ctx := context.Background()
	principal := onboardingGuardPrincipal()

	t.Run("assignment first", func(t *testing.T) {
		store := NewDevelopmentAIAgentClientStore()
		agentID := onboardingGuardCreate(t, store, principal, "assign first", AgentVisibilityPrivate)
		if _, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-ns-1", AssignAIAgentTaskRequest{AgentID: agentID, RequestID: "req-ns"}); err != nil {
			t.Fatalf("assignment with request_id: %v", err)
		}
		// Same task, agent and request_id as the assignment: must not replay the assignment receipt.
		_, err := onboardingGuardComplete(store, principal, "task-ns-1", agentID, "req-ns")
		if !errors.Is(err, errOnboardingCompletionConflict) {
			t.Fatalf("completion sharing an assignment receipt error = %v, want conflict", err)
		}
	})

	t.Run("onboarding first", func(t *testing.T) {
		store := NewDevelopmentAIAgentClientStore()
		agentID := onboardingGuardCreate(t, store, principal, "onboarding first", AgentVisibilityPrivate)
		completed, err := onboardingGuardComplete(store, principal, "task-ns-2", agentID, "req-ns")
		if err != nil {
			t.Fatalf("completion: %v", err)
		}
		assigned, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-ns-2", AssignAIAgentTaskRequest{AgentID: agentID, RequestID: "req-ns"})
		if err != nil {
			t.Fatalf("assignment with the completion request_id: %v", err)
		}
		if assigned.AssignmentID == completed.AssignmentID || isOnboardingSeedAssignmentID(assigned.AssignmentID) {
			t.Fatalf("assignment replayed the onboarding completion: %+v", assigned)
		}
		if assigned.AssignmentState == AgentAssignmentStateCompleted {
			t.Fatalf("assignment took the completed onboarding state: %+v", assigned)
		}
	})
}

// CP4: a durable generation on the membership makes a seed-ID removal a conflict.
func TestOnboardingSeedRemovalRejectedWhenMembershipHasDurableGeneration(t *testing.T) {
	ctx := context.Background()
	principal := onboardingGuardPrincipal()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	persistent, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	agentID := onboardingGuardCreate(t, persistent.DevelopmentAIAgentClientStore, principal, "durable removal", AgentVisibilityPrivate)
	if _, err := persistent.CompleteAIAgentOnboardingTask(ctx, principal, "task-durable-removal", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     agentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	}); err != nil {
		t.Fatalf("CompleteAIAgentOnboardingTask: %v", err)
	}
	seedID := onboardingSeedAssignmentID("task-durable-removal", agentID)
	// Leave a durable generation on the membership, as a later publication would.
	persistent.mu.Lock()
	member := persistent.taskAgentMemberships["task-durable-removal"][agentID]
	member.DurableGeneration = "asn-later@1"
	persistent.taskAgentMemberships["task-durable-removal"][agentID] = member
	persistent.mu.Unlock()
	if err := persistent.saveSnapshot(ctx); err != nil {
		t.Fatalf("saveSnapshot: %v", err)
	}

	durable := NewStoreWithConfig(StoreConfig{AgentRegistry: persistent})
	defer durable.Close()
	server := NewServer(ServerConfig{AIAgentClient: persistent, Assignment: durable})
	if _, err := server.prepareAIAgentMembershipRemoval(ctx, principal, "task-durable-removal", agentID, seedID, "test removal"); !errors.Is(err, ErrAIAgentMembershipChanged) {
		t.Fatalf("seed-ID removal with durable generation error = %v, want membership changed", err)
	}
}

// Provenance and the one-time rule are persisted, so a restart cannot reset them.
func TestOnboardingCompletionStateSurvivesRestart(t *testing.T) {
	ctx := context.Background()
	principal := onboardingGuardPrincipal()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	first, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open first: %v", err)
	}
	agentID := onboardingGuardCreate(t, first.DevelopmentAIAgentClientStore, principal, "restart", AgentVisibilityPrivate)
	if _, err := first.CompleteAIAgentOnboardingTask(ctx, principal, "task-restart-1", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     agentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	}); err != nil {
		t.Fatalf("completion before restart: %v", err)
	}

	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	_, err = reopened.CompleteAIAgentOnboardingTask(ctx, principal, "task-restart-2", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     agentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	})
	if !errors.Is(err, errOnboardingCompletionConflict) {
		t.Fatalf("second completion after restart error = %v, want conflict (provenance and one-time rule kept)", err)
	}
	seeded, err := reopened.CompleteAIAgentOnboardingTask(ctx, principal, "task-restart-3", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     "agent-owned-codex",
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	})
	if !errors.Is(err, errOnboardingCompletionIneligible) {
		t.Fatalf("seeded agent after restart error = %v (response %+v), want ineligible", err, seeded)
	}
}
