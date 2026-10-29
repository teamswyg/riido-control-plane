package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func (f *coverageLifecycleFixture) poll(t *testing.T) PollResponse {
	t.Helper()
	body := PollRequest{DeviceID: f.enrollment.DeviceID, DaemonID: f.binding.DaemonID, RuntimeID: f.binding.RuntimeID, WaitMs: 1}
	response := f.post(t, "/v1/agents/"+f.binding.AgentID+"/poll", body, "device", "")
	var poll PollResponse
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &poll) != nil || poll.Action != PollStart || poll.Assignment == nil {
		t.Fatalf("actual covered claim status=%d", response.Code)
	}
	return poll
}

func (f *coverageLifecycleFixture) terminate(t *testing.T, poll PollResponse) {
	t.Helper()
	a := poll.Assignment
	body := AgentEventRequest{AssignmentID: a.ID, TaskID: a.TaskID, DeviceID: f.enrollment.DeviceID, DaemonID: f.binding.DaemonID, RuntimeID: f.binding.RuntimeID, RuntimeProvider: f.binding.RuntimeProvider, State: AssignmentFailed, EventType: EventAssignmentFailed, Message: "synthetic terminal"}
	response := f.post(t, "/v1/agents/"+f.binding.AgentID+"/events", body, "device", a.LeaseToken)
	if response.Code != http.StatusOK {
		t.Fatalf("actual terminal event status=%d", response.Code)
	}
}

func (f *coverageLifecycleFixture) prove(t *testing.T, poll PollResponse) string {
	t.Helper()
	body := map[string]any{"schema_version": "riido-execution-termination.v1", "kind": "local_invocations_quiescent", "started_invocations": 1}
	response := f.post(t, "/v1/daemon/execution-claims/"+poll.Assignment.ID+"/termination", body, "device", poll.Assignment.LeaseToken)
	if response.Code != http.StatusOK {
		t.Fatalf("actual termination receipt status=%d", response.Code)
	}
	return response.Body.String()
}
