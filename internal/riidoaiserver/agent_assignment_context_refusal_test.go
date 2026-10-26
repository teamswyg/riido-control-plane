package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAssignmentContextRefusals(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	if code, _ := f.read(AssignmentContextViewBase, "", "wrong-lease"); code != http.StatusForbidden {
		t.Fatalf("wrong lease status=%d", code)
	}
	if code, _ := f.read(AssignmentContextViewTask, "RIID-1", f.leased.LeaseToken); code != http.StatusForbidden {
		t.Fatalf("out-of-snapshot target status=%d", code)
	}
	reads, _ := f.contexts.ListAssignmentContextReads(context.Background(), f.leased.ID)
	if last := reads[len(reads)-1]; last.Status != AssignmentContextReadDenied || last.SourceID != "RIID-1" {
		t.Fatalf("denied read = %+v", last)
	}
	body, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken})
	if code, _ := f.do("/v1/agents/agent-owned-codex/assignment-context", string(body), func(r *http.Request) { r.Header.Set("Authorization", "Bearer user-token") }); code != http.StatusUnauthorized {
		t.Fatalf("user bearer must not read snapshots, status=%d", code)
	}
	if code, _ := f.do("/v1/agents/agent-owned-codex/assignment-context", string(body), func(r *http.Request) {
		r.Header.Set(deviceIDHeader, "device-dev-macbook")
		r.Header.Set(deviceSecretHeader, "wrong")
	}); code != http.StatusUnauthorized {
		t.Fatalf("wrong device secret status=%d", code)
	}
	if code, _ := f.do("/v1/agents/agent-other/assignment-context", string(body), f.device); code != http.StatusNotFound {
		t.Fatalf("other agent path status=%d", code)
	}
}

func TestAssignmentContextEndsWithTheRun(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	for _, phase := range []struct {
		state AssignmentState
		kind  string
	}{{AssignmentReady, EventAssignmentReady}, {AssignmentRunning, EventAssignmentRunning}} {
		f.event(phase.state, phase.kind)
		if code, body := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", phase.state, code, body)
		}
	}
	f.event(AssignmentCompleted, EventAssignmentCompleted)
	if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusConflict {
		t.Fatalf("completed run status=%d", code)
	}
}

func (f *assignmentContextFixture) event(state AssignmentState, kind string) {
	body, _ := json.Marshal(AgentEventRequest{AssignmentID: f.leased.ID, TaskID: f.leased.TaskID, DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", State: state, EventType: kind})
	if code, resp := f.do("/v1/agents/agent-owned-codex/events", string(body), f.device); code != http.StatusOK {
		f.t.Fatalf("event %s status=%d body=%s", kind, code, resp)
	}
}
