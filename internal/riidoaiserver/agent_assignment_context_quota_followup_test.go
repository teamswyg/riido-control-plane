package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAssignmentContextQuotaNewFollowupGetsIndependentBudget(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	for range 100 {
		if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusOK {
			t.Fatal(code)
		}
	}
	f.event(AssignmentReady, EventAssignmentReady)
	// A continued native run must have a real provider session to resume.
	pin, _ := json.Marshal(AgentEventRequest{AssignmentID: f.leased.ID, TaskID: f.leased.TaskID, DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", State: AssignmentRunning, EventType: EventProviderSessionPinned, ProviderSessionID: "quota-native-session"})
	if code, body := f.do("/v1/agents/agent-owned-codex/events", string(pin), f.device); code != http.StatusOK {
		t.Fatalf("native session pin: %d %s", code, body)
	}
	f.event(AssignmentRunning, EventAssignmentRunning)
	f.event(AssignmentCompleted, EventAssignmentCompleted)
	old := f.leased
	// Physical TTL removal cannot revive a terminal capability.
	f.contexts.mu.Lock()
	delete(f.contexts.readAttempts, old.ID)
	f.contexts.mu.Unlock()
	if code, _ := f.read(AssignmentContextViewBase, "", old.LeaseToken); code != http.StatusConflict {
		t.Fatalf("terminal capability status=%d", code)
	}
	f.agents.mu.Lock()
	threadID := ""
	for _, thread := range f.agents.taskThreads[old.TaskID] {
		if thread.AssignmentID == old.ID {
			threadID = thread.ThreadID
		}
	}
	f.agents.mu.Unlock()
	if threadID == "" {
		t.Fatal("missing first thread")
	}
	path := "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/" + old.TaskID + "/threads/" + threadID + "/messages"
	code, body := f.do(path, `{"body":"Continue with another step","request_id":"followup-quota"}`, func(r *http.Request) { r.Header.Set(aiAgentTokenHeader, "user-token") })
	var next AIAgentTaskActionResponse
	if code != http.StatusAccepted || json.Unmarshal([]byte(body), &next) != nil || next.AssignmentID == old.ID || next.AssignmentID == "" {
		t.Fatalf("followup status=%d body=%s", code, body)
	}
	poll, err := f.service.assignment.(*Store).PollAgent(t.Context(), old.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
	if err != nil || poll.Assignment == nil || poll.Assignment.ID != next.AssignmentID {
		t.Fatalf("followup poll=%+v err=%v", poll, err)
	}
	f.leased = *poll.Assignment
	for i := range 100 {
		if code, body := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusOK {
			t.Fatalf("new execution attempt %d status=%d body=%s", i+1, code, body)
		}
	}
	if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusTooManyRequests {
		t.Fatalf("new execution attempt101=%d", code)
	}
}
