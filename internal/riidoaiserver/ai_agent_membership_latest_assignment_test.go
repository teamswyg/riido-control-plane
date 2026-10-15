package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
)

func TestMembershipLatestAssignmentObservesFollowupAfterWaitingRoot(t *testing.T) {
	for _, state := range []AssignmentState{AssignmentRunning, AssignmentCompleted, AssignmentFailed} {
		t.Run(string(state), func(t *testing.T) {
			ctx := context.Background()
			store := NewDevelopmentAIAgentClientStore()
			p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
			root := assignConversationThread(t, store, ctx, p, "task-latest-membership", "asn-root")
			recordNeedsInput(t, store, root, "Please specify the next action")
			followup, err := store.CreateAIAgentTaskThreadMessage(ctx, p, root.TaskID, root.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "implement the next action", AssignmentID: "asn-followup", durableState: state})
			if err != nil {
				t.Fatal(err)
			}
			response, err := store.ListWorkspaceAssignedAgentProfiles(ctx, p)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(response.TaskAgentParticipants[root.TaskID])
			if err != nil {
				t.Fatal(err)
			}
			var wire struct {
				Agents []struct {
					LatestAssignmentID string `json:"latest_assignment_id"`
				} `json:"agents"`
			}
			if err := json.Unmarshal(raw, &wire); err != nil {
				t.Fatal(err)
			}
			if len(wire.Agents) != 1 || wire.Agents[0].LatestAssignmentID != followup.AssignmentID {
				t.Fatalf("latest assignment must observe followup, independently of state: %s", raw)
			}
		})
	}
}

func TestMembershipWithoutRetainedHistoryOmitsLatestAssignment(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.mu.Lock()
	store.assignTaskAgentMemberLocked("task-no-history", store.agents["agent-owned-codex"])
	store.mu.Unlock()
	response, err := store.ListWorkspaceAssignedAgentProfiles(context.Background(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(response.TaskAgentParticipants["task-no-history"])
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Agents []map[string]any `json:"agents"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	if len(wire.Agents) != 1 {
		t.Fatalf("participants=%s", raw)
	}
	if _, exists := wire.Agents[0]["latest_assignment_id"]; exists {
		t.Fatalf("no-history ID must be absent: %s", raw)
	}
}
