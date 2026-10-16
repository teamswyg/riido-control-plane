package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

// A padded agent_id must be normalized before the scoped membership commit. The
// commit keys its delta by the agent id, so a padded id would drop the trimmed
// participant from the committed relation. A fresh replica must see the participant.
func TestDynamoDBOnboardingCompletionTrimsPaddedAgentIDAcrossReplicas(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	f := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, f)
	created, err := writer.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("padded id"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	agentID := created.Agent.AgentID

	response, err := writer.CompleteAIAgentOnboardingTask(ctx, principal, "  task-padded-dyn  ", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     "  " + agentID + "\t",
		TemplateKey: AIAgentOnboardingTemplateSchedule,
	})
	if err != nil {
		t.Fatalf("padded completion: %v", err)
	}
	if response.AgentID != agentID || response.TaskID != "task-padded-dyn" {
		t.Fatalf("response ids = %q/%q, want %q/task-padded-dyn", response.AgentID, response.TaskID, agentID)
	}

	reader, _ := newMembershipDynamoReplica(t, f)
	reader.mu.Lock()
	defer reader.mu.Unlock()
	if _, ok := reader.taskAgentMemberships["task-padded-dyn"][agentID]; !ok {
		t.Fatalf("fresh replica lost the participant; memberships = %v", reader.taskAgentMemberships)
	}
	for taskID, members := range reader.taskAgentMemberships {
		for key := range members {
			if strings.TrimSpace(key) != key || strings.TrimSpace(taskID) != taskID {
				t.Fatalf("padded membership key %q/%q was committed", taskID, key)
			}
		}
	}
}
