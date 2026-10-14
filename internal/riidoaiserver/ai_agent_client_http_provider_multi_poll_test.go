package riidoaiserver

import (
	"net/http"
	"testing"
)

func assertProviderMultiPollStart(t *testing.T, server http.Handler, token string, assigned AIAgentTaskActionResponse, daemonID, deviceID, runtimeID string) {
	t.Helper()
	body := `{"daemon_id":"` + daemonID + `","device_id":"` + deviceID + `","runtime_id":"` + runtimeID + `"}`
	bytes := aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/agents/"+assigned.AgentID+"/poll", token, body, http.StatusOK)
	var poll PollResponse
	aiAgentSmokeDecode(t, bytes, &poll)
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.ID != assigned.AssignmentID {
		t.Fatalf("poll start for %s = %+v", assigned.AssignmentID, poll)
	}
}
