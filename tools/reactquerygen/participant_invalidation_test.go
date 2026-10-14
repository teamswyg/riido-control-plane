package main

import (
	"strings"
	"testing"
)

func TestTaskParticipantMutationsInvalidateWorkspaceParticipants(t *testing.T) {
	spec := loadTestOpenAPI(t)
	for _, id := range []string{"createAIAgentTaskAgentAssignmentV2", "deleteAIAgentTaskAgentAssignmentV2", "stopAIAgentTaskAgentAssignmentV2", "createAIAgentTaskThreadMessageV2", "assignAIAgentTaskV2", "unassignAIAgentTaskV2", "stopAIAgentTaskV2"} {
		found := false
		for _, methods := range spec.Paths {
			for _, op := range methods {
				if op.OperationID != id {
					continue
				}
				for _, tag := range op.Client.Invalidates {
					if tag == "v2.aiAgent.tasks.assignedAgentProfiles" {
						found = true
					}
				}
			}
		}
		if !found {
			t.Fatalf("%s leaves workspace participants stale", id)
		}
	}
	body, err := generate(spec)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(body), "aiAgentTasksAssignedAgentProfiles: (queryClient: QueryClient)") < 7 {
		t.Fatal("generated mutation helpers omit participant invalidation")
	}
}
