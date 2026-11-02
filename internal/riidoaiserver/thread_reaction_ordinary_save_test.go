package riidoaiserver

import "testing"

func TestOrdinaryReplyPreservesReactionCommittedByAnotherReplica(t *testing.T) {
	for _, remove := range []bool{false, true} {
		t.Run(map[bool]string{false: "add", true: "remove"}[remove], func(t *testing.T) {
			fixture := newMembershipSharedDynamoFixture(t)
			first, _ := newMembershipDynamoReplica(t, fixture)
			owner := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
			assigned, err := first.AssignAIAgentTask(t.Context(), owner, "task-replica", AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := first.CreateAIAgentTaskThreadMessage(t.Context(), owner, "task-replica", assigned.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "original"}); err != nil {
				t.Fatal(err)
			}
			threadID, messageID := replicaUserMessage(t, first, owner, "original")
			if remove {
				if _, err := first.MutateTaskMessageReaction(t.Context(), owner, "task-replica", threadID, messageID, "1f44d", false); err != nil {
					t.Fatal(err)
				}
			}
			second, _ := newMembershipDynamoReplica(t, fixture)
			if _, err := second.MutateTaskMessageReaction(t.Context(), owner, "task-replica", threadID, messageID, "1f44d", remove); err != nil {
				t.Fatal(err)
			}
			if !remove {
				for _, name := range []string{"renamed once", "renamed twice"} {
					if _, err := first.UpdateAIAgentConfiguration(t.Context(), owner, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: name}); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err := first.CreateAIAgentTaskThreadMessage(t.Context(), owner, "task-replica", threadID, CreateAIAgentTaskThreadMessageRequest{Body: "later"}); err != nil {
				t.Fatal(err)
			}
			reader, _ := newMembershipDynamoReplica(t, fixture)
			history, err := reader.ListAIAgentTaskThreadHistory(t.Context(), owner, "task-replica")
			if err != nil {
				t.Fatal(err)
			}
			found, laterCount := false, 0
			for _, thread := range history.Threads {
				for _, message := range thread.Messages {
					if message.Body == "later" {
						laterCount++
					}
					if message.MessageID == messageID {
						found = true
						if reacted := memberReactionEmoji(message.Reactions, owner.PrincipalID) != ""; reacted == remove {
							t.Fatalf("remove=%v, persisted reactions=%+v", remove, message.Reactions)
						}
					}
				}
			}
			if !found || laterCount != 1 {
				t.Fatalf("original kept=%v, later message count=%d", found, laterCount)
			}
		})
	}
}
