package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func assertDeliveredCommentContext(t *testing.T, assignments *Store, a AIAgentTaskActionResponse, evidence verifiedCommentDelivery) {
	t.Helper()
	poll, err := assignments.PollAgent(t.Context(), a.AgentID, PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
	if err != nil || poll.Action != PollStart || poll.Assignment == nil {
		t.Fatalf("verified comment did not reach provider: %+v %v", poll, err)
	}
	for _, want := range []string{evidence.Body, evidence.Context.Component.Title, evidence.Context.Document.Content} {
		if !strings.Contains(poll.Assignment.Prompt, want) {
			t.Fatalf("verified comment context lost: %q", want)
		}
	}
}

type retryableProjectionClient struct {
	*DevelopmentAIAgentClientStore
	fail bool
}

func (c *retryableProjectionClient) CreateAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if c.fail {
		c.fail = false
		return AIAgentTaskActionResponse{}, errors.New("snapshot temporarily unavailable")
	}
	return c.DevelopmentAIAgentClientStore.CreateAIAgentTaskAgentAssignment(ctx, principal, taskID, req)
}
