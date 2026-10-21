package riidoaiserver

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

// onboardingSplitItemsWithoutOnboardingParts models a snapshot written before the
// onboarding parts existed: those part items and their manifest hashes are absent.
func onboardingSplitItemsWithoutOnboardingParts(t *testing.T, items []map[string]map[string]string) []map[string]map[string]string {
	t.Helper()
	dropped := map[string]bool{
		dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartOnboardingCreated): true,
		dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartOnboardingDone):    true,
		dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartOnboardingThreads): true,
	}
	var out []map[string]map[string]string
	for _, item := range items {
		sk := item["sk"]["S"]
		if dropped[sk] {
			continue
		}
		if sk == dynamoDBAIAgentClientSnapshotSK {
			hashes := map[string]string{}
			if err := json.Unmarshal([]byte(item["part_hashes_json"]["S"]), &hashes); err != nil {
				t.Fatalf("manifest hashes: %v", err)
			}
			delete(hashes, dynamoDBAIAgentClientSnapshotPartOnboardingCreated)
			delete(hashes, dynamoDBAIAgentClientSnapshotPartOnboardingDone)
			delete(hashes, dynamoDBAIAgentClientSnapshotPartOnboardingThreads)
			copied := map[string]map[string]string{}
			for k, v := range item {
				copied[k] = v
			}
			copied["part_hashes_json"] = map[string]string{"S": mustJSONDynamoDBAIAgentClientSnapshotPartHashes(hashes)}
			item = copied
		}
		out = append(out, item)
	}
	return out
}

func TestDynamoDBSplitSnapshotRoundTripsOnboardingState(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	created, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("split onboarding"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	if _, err := store.CompleteAIAgentOnboardingTask(ctx, principal, "task-split-1", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     created.Agent.AgentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	}); err != nil {
		t.Fatalf("completion: %v", err)
	}
	snapshot, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.OnboardingCreatedIDs) != 1 || snapshot.OnboardingCompletions[created.Agent.AgentID] != "task-split-1" {
		t.Fatalf("snapshot onboarding state = %v / %v", snapshot.OnboardingCreatedIDs, snapshot.OnboardingCompletions)
	}

	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
	if err != nil {
		t.Fatalf("encode split snapshot: %v", err)
	}
	decoded, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(items)
	if err != nil {
		t.Fatalf("decode split snapshot: %v", err)
	}
	if !reflect.DeepEqual(decoded.OnboardingCreatedIDs, snapshot.OnboardingCreatedIDs) {
		t.Fatalf("created ids = %v, want %v", decoded.OnboardingCreatedIDs, snapshot.OnboardingCreatedIDs)
	}
	if !reflect.DeepEqual(decoded.OnboardingCompletions, snapshot.OnboardingCompletions) {
		t.Fatalf("completions = %v, want %v", decoded.OnboardingCompletions, snapshot.OnboardingCompletions)
	}
	if len(decoded.OnboardingCompletionThreads) != 1 || decoded.OnboardingCompletionThreads[created.Agent.AgentID].TaskID != "task-split-1" {
		t.Fatalf("completion threads = %+v, want one thread for task-split-1", decoded.OnboardingCompletionThreads)
	}

	// The restored store keeps provenance and the one-time rule.
	restored := NewDevelopmentAIAgentClientStore()
	if err := restored.restoreSnapshot(decoded); err != nil {
		t.Fatalf("restore: %v", err)
	}
	_, err = restored.CompleteAIAgentOnboardingTask(ctx, principal, "task-split-2", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     created.Agent.AgentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	})
	if !errorsIsConflict(err) {
		t.Fatalf("second completion after split round trip error = %v, want conflict", err)
	}
}

func TestDynamoDBSplitSnapshotWithoutOnboardingPartsDecodesEmpty(t *testing.T) {
	snapshot := NewDevelopmentAIAgentClientStore()
	current, err := snapshot.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(current, nil)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	legacy := onboardingSplitItemsWithoutOnboardingParts(t, items)
	decoded, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(legacy)
	if err != nil {
		t.Fatalf("decode legacy split snapshot: %v", err)
	}
	if len(decoded.OnboardingCreatedIDs) != 0 || len(decoded.OnboardingCompletions) != 0 {
		t.Fatalf("legacy snapshot onboarding state = %v / %v, want empty", decoded.OnboardingCreatedIDs, decoded.OnboardingCompletions)
	}
	if len(decoded.Agents) != len(current.Agents) {
		t.Fatalf("legacy snapshot agents = %d, want %d", len(decoded.Agents), len(current.Agents))
	}
}

func errorsIsConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "conflicts with existing agent history")
}
