package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestAssignmentContextQuotaSurvivesHeartbeatRunLabelAndRecapture(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	f.event(AssignmentReady, EventAssignmentReady)
	f.event(AssignmentRunning, EventAssignmentRunning)
	for i := range 100 {
		if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusOK {
			t.Fatalf("attempt %d status=%d", i+1, code)
		}
	}
	heartbeat, err := f.service.assignment.(*Store).HeartbeatAgent(t.Context(), f.leased.AgentID, AgentHeartbeatRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", ActiveAssignmentIDs: []string{f.leased.ID}})
	if err != nil || len(heartbeat.RefreshedAssignments) != 1 {
		t.Fatalf("heartbeat=%+v err=%v", heartbeat, err)
	}
	if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusTooManyRequests {
		t.Fatalf("heartbeat reset quota: %d", code)
	}
	progress, err := f.agents.RecordAIAgentThreadProgress(t.Context(), f.leased.AgentID, AgentThreadProgressBatchRequest{AssignmentID: f.leased.ID, TaskID: f.leased.TaskID, RunID: "replacement-runtime-label", Lines: []AgentThreadProgressLine{{Seq: 500, Message: "Still working"}}})
	if err != nil || progress.Event.RunID != "replacement-runtime-label" {
		t.Fatalf("progress=%+v err=%v", progress, err)
	}
	if code, _ := f.read(AssignmentContextViewAgents, "", f.leased.LeaseToken); code != http.StatusTooManyRequests {
		t.Fatalf("runtime run label reset quota: %d", code)
	}
	f.captureForTest(t, f.leased)
	if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusTooManyRequests {
		t.Fatalf("recapture reset quota: %d", code)
	}
}
