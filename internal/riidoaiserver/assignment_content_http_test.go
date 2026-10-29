package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAssignmentContentHTTPReadsOnlyCapturedPage(t *testing.T) {
	f := newAssignmentContentHTTPFixture(t)
	for _, test := range []struct {
		target, cursor string
		status         int
	}{
		{"task-private", "first", http.StatusOK},
		{"task-private", "last", http.StatusOK},
		{"another", "first", http.StatusForbidden},
		{"task-private", "invented", http.StatusForbidden},
	} {
		request, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: "content", Target: test.target, Cursor: test.cursor})
		status, body := f.do("/v1/agents/agent-owned-codex/assignment-context", string(request), f.device)
		if status != test.status {
			t.Fatalf("%s/%s status=%d body=%s", test.target, test.cursor, status, body)
		}
		if status == http.StatusOK && !strings.Contains(body, `"content":"ab"`) {
			t.Fatalf("missing captured page: %s", body)
		}
	}
	reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
	if err != nil {
		t.Fatal(err)
	}
	served := 0
	for _, read := range reads {
		if read.Via == AssignmentContextViaCLI && read.Status == AssignmentContextReadServed && read.SourceID == "task-private" && read.SourceVersion == "v1" {
			served++
		}
	}
	if served != 2 {
		t.Fatalf("captured page audit count=%d", served)
	}
}

func newAssignmentContentHTTPFixture(t *testing.T) *assignmentContextFixture {
	t.Helper()
	m := NewMemoryAssignmentContextStore()
	f := newAssignmentContextFixture(t, m)
	snapshot, _, err := m.LoadAssignmentContextSnapshot(t.Context(), f.leased.ID)
	if err != nil {
		t.Fatal(err)
	}
	content, pages := assignmentContentTestBundle()
	content.Content[0].SourceID = "task-private"
	for i := range pages {
		pages[i].SourceID, pages[i].Sources[0].ID = "task-private", "task-private"
	}
	snapshot.Content = content.Content
	// Replace only this synthetic legacy capture to isolate the new HTTP reader.
	m.mu.Lock()
	delete(m.snapshots, f.leased.ID)
	m.mu.Unlock()
	if err := m.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	return f
}
