package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAssignmentContextReadsStreamToClients(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	if code, body := f.read(AssignmentContextViewTask, "RIID-4800", f.leased.LeaseToken); code != http.StatusOK {
		t.Fatalf("read status=%d body=%s", code, body)
	}
	f.read(AssignmentContextViewTask, "RIID-1", f.leased.LeaseToken)
	events := f.contextReadEvents("user-token")
	if len(events) != 2 {
		t.Fatalf("want prompt and served events; out-of-scope denied source must stay hidden, got %+v", events)
	}
	prompt, served := events[0], events[1]
	if prompt.Via != AssignmentContextViaPrompt || len(prompt.Sources) != 4 || prompt.ThreadID == "" || prompt.RunID == "" || prompt.AssignmentID != f.leased.ID || prompt.TriggerKind != AssignmentContextTriggerAssign {
		t.Fatalf("prompt event = %+v", prompt)
	}
	if served.Via != AssignmentContextViaCLI || served.Sources[0].Kind != "project" || served.Sources[0].UpdatedAt != "2026-09-30T15:00:00Z" || served.Sources[0].Version == "" || served.Sources[0].Status != AssignmentContextReadServed {
		t.Fatalf("served event = %+v", served)
	}
	reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
	if err != nil || reads[len(reads)-1].Status != AssignmentContextReadDenied {
		t.Fatal("hidden denied reads must remain in the server audit")
	}
	if encoded, _ := json.Marshal(events); strings.Contains(string(encoded), "user-1") || strings.Contains(string(encoded), "cli_1") {
		t.Fatalf("requester and CLI request ids must stay server-side: %s", encoded)
	}
	if other := f.contextReadEvents("other-token"); len(other) != 0 {
		t.Fatalf("a private agent's reads leaked to another member: %+v", other)
	}
}

func TestAssignmentContextViewsStayOnTheAssignedTask(t *testing.T) {
	c := assignmentContextTestTaskContext()
	for _, tc := range []struct {
		view, target string
		ok           bool
	}{{AssignmentContextViewTree, "RIID-5900", true}, {AssignmentContextViewTree, "RIID-1", false}, {AssignmentContextViewResources, "task-private", true}, {AssignmentContextViewBase, "RIID-4800", false}, {AssignmentContextViewTask, "milestone-a", true}} {
		if _, _, ok := selectAssignmentContextView(c, tc.view, tc.target); ok != tc.ok {
			t.Fatalf("%s %s ok=%v", tc.view, tc.target, ok)
		}
	}
}
