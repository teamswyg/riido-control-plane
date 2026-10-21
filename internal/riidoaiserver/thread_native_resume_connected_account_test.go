package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPConnectedAccountAgentCanContinueItsOwnNativeSession(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	ctx := t.Context()
	owner := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	guest := AuthorizationResult{PrincipalID: "user-2", WorkspaceID: defaultAIAgentClientWorkspaceID}
	const machineID = "native-shared-machine"
	enrollment := enrollTestDevice(t, ctx, f.client, owner, owner.WorkspaceID, machineID, "Shared device")
	runtimeID := enrollment.DeviceID + ":claude"
	syncTestRuntime(t, ctx, f.client, owner, machineID, enrollment.DeviceID, runtimeID)
	if _, err := f.client.ConnectAIAgentDevice(ctx, guest, machineID); err != nil {
		t.Fatal(err)
	}
	created, err := f.client.CreateAIAgent(ctx, guest, CreateAgentConfigurationRequest{Name: "Guest Agent", RuntimeID: runtimeID, Visibility: AgentVisibilityPrivate})
	if err != nil {
		t.Fatal(err)
	}
	_, runtime, found := f.client.LookupAgentRuntimeFact(created.Agent.AgentID)
	if !found || runtime.OwnerPrincipalID == created.Agent.OwnerPrincipalID {
		t.Fatal("fixture must preserve the device owner separately from the connected account Agent owner")
	}
	f.token = "native-teammate-token"
	var current AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{AgentID: created.Agent.AgentID}, http.StatusAccepted, &current)
	for round := range 3 {
		polled, err := f.store.PollAgent(ctx, created.Agent.AgentID, PollRequest{DaemonID: machineID, DeviceID: enrollment.DeviceID, RuntimeID: runtimeID})
		if err != nil || polled.Assignment == nil || polled.Assignment.ID != current.AssignmentID {
			t.Fatal("assignment was not delivered to the original physical daemon")
		}
		if round > 0 && polled.Assignment.ResumeSessionID != "connected-account-native-session" {
			t.Fatal("followup lost its own selected native session")
		}
		for _, phase := range []struct {
			state     AssignmentState
			kind, pin string
		}{{AssignmentReady, EventAssignmentReady, ""}, {AssignmentRunning, EventProviderSessionPinned, "connected-account-native-session"}, {AssignmentCompleted, EventAssignmentCompleted, "connected-account-native-session"}} {
			f.post(t, "/v1/agents/"+created.Agent.AgentID+"/events", AgentEventRequest{AssignmentID: current.AssignmentID, TaskID: "task-new", DaemonID: machineID, DeviceID: enrollment.DeviceID, RuntimeID: runtimeID, State: phase.state, EventType: phase.kind, ProviderSessionID: phase.pin}, http.StatusOK, nil)
		}
		if round < 2 {
			f.post(t, f.base+"/tasks/task-new/threads/"+current.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "Continue the same task."}, http.StatusAccepted, &current)
		}
	}
	// Device ownership does not authorize another account to invoke the private Agent.
	f.token = "native-owner-token"
	f.post(t, f.base+"/tasks/task-new/threads/"+current.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "Not my private Agent."}, http.StatusNotFound, nil)
}
