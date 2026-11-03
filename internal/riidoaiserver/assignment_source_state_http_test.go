package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

// sourceStateHTTPFixture stores a previous capture with one extra related
// source and a current capture that recorded it as deleted.
func sourceStateHTTPFixture(t *testing.T) (*assignmentContextFixture, AssignmentContextSnapshot) {
	t.Helper()
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	current, found, err := f.contexts.LoadAssignmentContextSnapshot(t.Context(), f.leased.ID)
	if err != nil || !found {
		t.Fatal("missing initial snapshot", err)
	}
	current.Context.TeamID, current.ThreadID = "fixture-team", "current-thread"
	previous := current
	previous.AssignmentID, previous.ThreadID, previous.ConversationID = "previous", "previous-thread", "previous-thread"
	current.ConversationID, current.PreviousAssignmentID, current.PreviousThreadID = previous.ConversationID, previous.AssignmentID, previous.ThreadID
	previous, pages := contextChangesTestBundle(t, previous, "OLD-PRIVATE-BODY")
	gone, gonePages := contextChangesTestBundle(t, AssignmentContextSnapshot{TaskID: "gone-work", Context: AIAgentTaskContext{TeamID: "fixture-team"}}, "GONE-PRIVATE-BODY")
	previous.Context.ReadScope.Resources = append(previous.Context.ReadScope.Resources, gone.Context.ReadScope.Resources[0])
	previous.Content = append(previous.Content, gone.Content...)
	if err = f.contexts.SaveAssignmentContentBundle(t.Context(), previous, append(pages, gonePages...)); err != nil {
		t.Fatal(err)
	}
	current, pages = contextChangesTestBundle(t, current, "NEW-PRIVATE-BODY")
	current.SourceStates = []AssignmentSourceState{{ResourceKind: AssignmentContentComponent, SourceID: "gone-work", State: AssignmentSourceDeleted}}
	f.contexts.mu.Lock()
	delete(f.contexts.snapshots, current.AssignmentID)
	f.contexts.mu.Unlock()
	if err = f.contexts.SaveAssignmentContentBundle(t.Context(), current, pages); err != nil {
		t.Fatal(err)
	}
	return f, previous
}

func TestContextChangesHTTPReportsSourceStatesOnlyToVersionTwo(t *testing.T) {
	f, previous := sourceStateHTTPFixture(t)
	for version, want := range map[string]string{"": "unavailable", "1": "unavailable", "2": "deleted"} {
		code, body := f.do(contextChangesTestPath, contextChangesRequest(f, previous.AssignmentID, f.leased.LeaseToken), func(r *http.Request) {
			f.device(r)
			if version != "" {
				r.Header.Set(AssignmentContextChangesVersionHeader, version)
			}
		})
		var response struct{ Data AssignmentContextChanges }
		if code != http.StatusOK || json.Unmarshal([]byte(body), &response) != nil {
			t.Fatalf("version %q: status=%d body=%s", version, code, body)
		}
		var status string
		for _, change := range response.Data.Changes {
			if change.SourceID == "gone-work" {
				status = change.Status
			}
		}
		if status != want || response.Data.FullRefreshRequired != (want == "unavailable") {
			t.Fatalf("version %q: gone-work=%q body=%s", version, status, body)
		}
	}
}
