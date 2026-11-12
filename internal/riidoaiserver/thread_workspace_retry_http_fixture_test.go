package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
)

func workspaceRetryHTTPFixture(t *testing.T, reconcileFirst bool) (*nativeResumeHTTPFixture, AIAgentTaskActionResponse, PollRequest) {
	t.Helper()
	f := newNativeResumeHTTPFixture(t)
	f.taskContext.contextSnapshot.Repositories[0].Source = "pull_request"
	f.taskContext.contextSnapshot.Component.KeyNumber = "RIID-6101"
	f.taskContext.contextSnapshot.Component.BranchName = "RIID-6101-missing-task-branch"
	repository := f.taskContext.contextSnapshot.Repositories[0]
	f.taskContext.contextSnapshot.linkedPullRequests = taskCheckoutPullRequests([]json.RawMessage{
		json.RawMessage(`{"number":1396,"htmlUrl":"https://github.com/` + repository.FullName + `/pull/1396"}`),
	})

	var first AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex", RequestID: "first-missing-branch", Instruction: "Verify the shortcut implementation.",
	}, http.StatusAccepted, &first)
	pollRequest := PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}
	poll, err := f.store.PollAgent(t.Context(), first.AgentID, pollRequest)
	if err != nil || poll.Assignment == nil || poll.Assignment.Worktree.Source != "pull_request_fallback:1396" {
		t.Fatalf("assignment did not include its sole linked PR: %+v, %v", poll.Assignment, err)
	}
	failure := AgentEventRequest{
		AssignmentID: first.AssignmentID, TaskID: first.TaskID, DaemonID: pollRequest.DaemonID, DeviceID: pollRequest.DeviceID, RuntimeID: pollRequest.RuntimeID,
		State: AssignmentFailed, EventType: EventAssignmentFailed,
		Metadata: map[string]string{assignment.ExecutionDiagnosticKey: "workspace_preparation_failed", assignment.ExecutionNextActionKey: "check_workspace"},
	}
	if reconcileFirst {
		// The durable assignment can become terminal before its client event
		// arrives, including when the event is delivered on another instance.
		result, err := f.store.RecordAgentEvent(t.Context(), first.AgentID, failure)
		if err != nil {
			t.Fatal(err)
		}
		principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
		if _, err := f.client.ReconcileAIAgentActiveThreadProjections(t.Context(), principal, first.TaskID, f.store); err != nil {
			t.Fatal(err)
		}
		if err := f.client.RecordAIAgentAssignmentEvent(t.Context(), first.AgentID, failure, result.Event); err != nil {
			t.Fatal(err)
		}
	} else {
		f.post(t, "/v1/agents/"+first.AgentID+"/events", failure, http.StatusOK, nil)
	}

	return f, first, pollRequest
}
