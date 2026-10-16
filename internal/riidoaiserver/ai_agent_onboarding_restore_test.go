package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

// legacyOnboardingSnapshot drops the onboarding parts that snapshots written before
// onboarding_created_agent_ids existed did not contain.
func legacyOnboardingSnapshot(snapshot AIAgentClientSnapshot) AIAgentClientSnapshot {
	snapshot.OnboardingCreatedIDs = nil
	snapshot.OnboardingCompletions = nil
	snapshot.OnboardingCompletionThreads = nil
	return snapshot
}

// reloadSnapshotThroughSplitFormat round-trips a snapshot through the DynamoDB split
// encoding, the same path a reload takes.
func reloadSnapshotThroughSplitFormat(t *testing.T, snapshot AIAgentClientSnapshot) AIAgentClientSnapshot {
	t.Helper()
	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
	if err != nil {
		t.Fatalf("encode split snapshot: %v", err)
	}
	decoded, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(items)
	if err != nil {
		t.Fatalf("decode split snapshot: %v", err)
	}
	return decoded
}

// appendFillerTaskThreads adds completed threads after the existing ones. Once the
// task has more than the snapshot limit, the earlier completion thread is dropped
// from the persisted task thread list.
func appendFillerTaskThreads(store *DevelopmentAIAgentClientStore, taskID string, count int) {
	store.mu.Lock()
	defer store.mu.Unlock()
	for i := range count {
		store.taskThreads[taskID] = append(store.taskThreads[taskID], AIAgentTaskThreadRecord{
			ThreadID:        fmt.Sprintf("thread-filler-%03d", i),
			TaskID:          taskID,
			AgentID:         "agent-filler",
			AssignmentID:    fmt.Sprintf("assignment-filler-%03d", i),
			WorkStatus:      AgentWorkStatusCompleted,
			AssignmentState: AgentAssignmentStateCompleted,
			CommentKind:     AgentTaskCommentTaskCompleted,
			Message:         "filler",
		})
	}
}

func persistedTaskThreadExists(snapshot AIAgentClientSnapshot, taskID, assignmentID string) bool {
	for _, thread := range snapshot.TaskThreads[taskID] {
		if thread.AssignmentID == assignmentID {
			return true
		}
	}
	return false
}

func assertSameJSON(t *testing.T, want, got any) {
	t.Helper()
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal want: %v", err)
	}
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal got: %v", err)
	}
	if string(wantJSON) != string(gotJSON) {
		t.Fatalf("response mismatch\nwant %s\n got %s", wantJSON, gotJSON)
	}
}

func TestOnboardingRestoreFromLegacySnapshotKeepsFixtureAgentEligible(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	fixtureAgent, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("legacy fixture"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	normalAgent, err := store.CreateAIAgent(ctx, principal, profileColorTestRequest("legacy normal"))
	if err != nil {
		t.Fatalf("normal create: %v", err)
	}
	snapshot, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	restored := NewDevelopmentAIAgentClientStore()
	if err := restored.restoreSnapshot(legacyOnboardingSnapshot(snapshot)); err != nil {
		t.Fatalf("restore legacy snapshot: %v", err)
	}

	if _, err := restored.CompleteAIAgentOnboardingTask(ctx, principal, "task-legacy-fixture", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     fixtureAgent.Agent.AgentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	}); err != nil {
		t.Fatalf("fixture-created agent completion after legacy restore: %v", err)
	}
	_, err = restored.CompleteAIAgentOnboardingTask(ctx, principal, "task-legacy-normal", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     normalAgent.Agent.AgentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	})
	if !errors.Is(err, errOnboardingCompletionIneligible) {
		t.Fatalf("normal agent completion after legacy restore error = %v, want ineligible", err)
	}
}

func TestOnboardingRestoreFromLegacySnapshotIgnoresNonFixtureTmpColor(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	normalAgent, err := store.CreateAIAgent(ctx, principal, profileColorTestRequest("legacy tinted"))
	if err != nil {
		t.Fatalf("normal create: %v", err)
	}
	snapshot, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	legacy := legacyOnboardingSnapshot(snapshot)
	for i := range legacy.Agents {
		if legacy.Agents[i].AgentID == normalAgent.Agent.AgentID {
			// A color outside the fixture table is not proof of fixture provenance.
			legacy.Agents[i].TmpColor = "#94A3B8"
		}
	}

	restored := NewDevelopmentAIAgentClientStore()
	if err := restored.restoreSnapshot(legacy); err != nil {
		t.Fatalf("restore legacy snapshot: %v", err)
	}
	_, err = restored.CompleteAIAgentOnboardingTask(ctx, principal, "task-legacy-tinted", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     normalAgent.Agent.AgentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	})
	if !errors.Is(err, errOnboardingCompletionIneligible) {
		t.Fatalf("non-fixture tmp_color completion after legacy restore error = %v, want ineligible", err)
	}
}

func TestOnboardingCompletionRepeatSurvivesTruncationAndReload(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	created, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("truncated repeat"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	const taskID = "task-truncated-1"
	request := CompleteAIAgentOnboardingTaskRequest{
		AgentID:     created.Agent.AgentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	}
	first, err := store.CompleteAIAgentOnboardingTask(ctx, principal, taskID, request)
	if err != nil {
		t.Fatalf("first completion: %v", err)
	}
	appendFillerTaskThreads(store, taskID, aiAgentClientThreadsPerTaskLimit+1)

	snapshot, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if persistedTaskThreadExists(snapshot, taskID, first.AssignmentID) {
		t.Fatalf("completion thread should be truncated out of the persisted task threads")
	}

	restored := NewDevelopmentAIAgentClientStore()
	if err := restored.restoreSnapshot(reloadSnapshotThroughSplitFormat(t, snapshot)); err != nil {
		t.Fatalf("restore after reload: %v", err)
	}
	replay, err := restored.CompleteAIAgentOnboardingTask(ctx, principal, taskID, request)
	if err != nil {
		t.Fatalf("repeat completion after truncation and reload: %v", err)
	}
	assertSameJSON(t, first, replay)

	_, err = restored.CompleteAIAgentOnboardingTask(ctx, principal, "task-truncated-2", request)
	if !errors.Is(err, errOnboardingCompletionConflict) {
		t.Fatalf("completion for another task error = %v, want conflict", err)
	}
	_, err = restored.StopAIAgentTask(ctx, principal, taskID, StopAIAgentTaskRequest{AgentID: created.Agent.AgentID})
	if !errors.Is(err, errOnboardingCompletionStopConflict) {
		t.Fatalf("agent-only stop after truncation error = %v, want stop conflict", err)
	}
}

func TestOnboardingCompletionRepeatWithoutStoredThreadConflictsAfterTruncation(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	created, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("old record"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	const taskID = "task-old-record-1"
	request := CompleteAIAgentOnboardingTaskRequest{
		AgentID:     created.Agent.AgentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	}
	if _, err := store.CompleteAIAgentOnboardingTask(ctx, principal, taskID, request); err != nil {
		t.Fatalf("first completion: %v", err)
	}
	appendFillerTaskThreads(store, taskID, aiAgentClientThreadsPerTaskLimit+1)
	snapshot, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	// A record written before completion threads were persisted has no thread copy,
	// so the repeat cannot be reproduced and must stay a conflict.
	snapshot.OnboardingCompletionThreads = nil

	restored := NewDevelopmentAIAgentClientStore()
	if err := restored.restoreSnapshot(snapshot); err != nil {
		t.Fatalf("restore: %v", err)
	}
	_, err = restored.CompleteAIAgentOnboardingTask(ctx, principal, taskID, request)
	if !errors.Is(err, errOnboardingCompletionConflict) {
		t.Fatalf("repeat without stored thread error = %v, want conflict", err)
	}
}
