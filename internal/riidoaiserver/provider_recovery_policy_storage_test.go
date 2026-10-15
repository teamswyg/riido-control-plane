package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestRecoveryPolicyPersistedCopiedAndRequestCannotChangeAuthority(t *testing.T) {
	store := NewStore()
	defer store.Close()
	policy := recoveryTestPolicy()
	req := AssignRequest{RequestID: "same-request", WorkspaceID: "workspace", CreatedBy: "owner", AgentID: "agent", ComponentID: "task", RuntimeProvider: "codex", Prompt: "run", RecoveryPolicy: policy}
	created, err := store.AssignTaskAdditive(context.Background(), "task", req)
	if err != nil {
		t.Fatal(err)
	}
	policy.Fallbacks[0].ModelID = "changed-input"
	created.RecoveryPolicy.Fallbacks[0].ModelID = "changed-response"
	projection, found, err := store.LoadAssignmentProjection(context.Background(), created.ID)
	if err != nil || !found || projection.Assignment.RecoveryPolicy.Fallbacks[0].ModelID != "cursor-auto" {
		t.Fatalf("policy was aliased: %+v %v", projection, err)
	}
	data, err := json.Marshal(projection.Assignment)
	if err != nil {
		t.Fatal(err)
	}
	var restored Assignment
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if !sameRecoveryPolicy(restored.RecoveryPolicy, recoveryTestPolicy()) {
		t.Fatal("policy not persisted")
	}
	req.RecoveryPolicy = recoveryTestPolicy()
	req.RecoveryPolicy.MaxAttempts = 2
	if _, err := store.AssignTaskAdditive(context.Background(), "task", req); !errors.Is(err, errRecoveryPolicyChanged) {
		t.Fatalf("changed RequestID policy accepted: %v", err)
	}
	req.RequestID = "pinned"
	req.ResumeSessionID = "native-session"
	if _, err := store.AssignTaskAdditive(context.Background(), "task", req); err == nil {
		t.Fatal("pinned cross-provider switch accepted")
	}
}
