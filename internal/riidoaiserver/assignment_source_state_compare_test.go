package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

// sourceStateTestBundle returns a capture of the root work item plus related
// sources, each with a private title and body.
func sourceStateTestBundle(t *testing.T, id string, extra ...string) (AssignmentContextSnapshot, []AssignmentContentPage) {
	t.Helper()
	snapshot, pages := contextChangesTestBundle(t, contextChangesTestSnapshot(id), "root body")
	for _, source := range extra {
		related, relatedPages := contextChangesTestBundle(t, AssignmentContextSnapshot{TaskID: source, Context: AIAgentTaskContext{TeamID: "team"}}, "PRIVATE-BODY-"+source)
		resource := related.Context.ReadScope.Resources[0]
		resource.Title, resource.Reasons = "PRIVATE-TITLE-"+source, []taskReadScopeReason{{Kind: "related"}}
		snapshot.Context.ReadScope.Resources = append(snapshot.Context.ReadScope.Resources, resource)
		snapshot.Content = append(snapshot.Content, related.Content...)
		pages = append(pages, relatedPages...)
	}
	return snapshot, pages
}

func TestContextChangesSourceStatesExplainOmissionsWithoutOldDetail(t *testing.T) {
	previous, _ := sourceStateTestBundle(t, "previous", "gone", "moved", "secret", "unknown")
	current, _ := sourceStateTestBundle(t, "current")
	current.SourceStates = []AssignmentSourceState{
		{ResourceKind: AssignmentContentComponent, SourceID: "gone", State: AssignmentSourceDeleted},
		{ResourceKind: AssignmentContentComponent, SourceID: "moved", State: AssignmentSourceUnselected},
		{ResourceKind: AssignmentContentComponent, SourceID: "secret", State: AssignmentSourceHidden},
	}
	out := compareAssignmentContextParts(previous, current)
	raw, _ := json.Marshal(out)
	got := map[string]string{}
	for _, change := range out.Changes {
		if change.Current != nil || change.Title != "" || change.TeamID != "" || len(change.Reasons) != 0 {
			t.Fatal("omitted source carried current metadata", string(raw))
		}
		got[change.SourceID] = change.Status + "/" + change.Reason
	}
	want := map[string]string{"gone": "deleted/", "moved": "unselected/", "unknown": "unavailable/not_in_current_capture"}
	for id, status := range want {
		if got[id] != status || len(got) != len(want) {
			t.Fatalf("%s = %q, want %q (%s)", id, got[id], status, raw)
		}
	}
	for _, forbidden := range []string{"PRIVATE", "secret", "hidden", "denied"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatal("result leaked an old detail or the hidden source", string(raw))
		}
	}
	if out.Complete || !out.FullRefreshRequired {
		t.Fatal("unknown omission no longer asks for a full read", string(raw))
	}
	current.SourceStates = append(current.SourceStates, AssignmentSourceState{ResourceKind: AssignmentContentComponent, SourceID: "unknown", State: AssignmentSourceDeleted})
	if out = compareAssignmentContextParts(previous, current); !out.Complete || out.FullRefreshRequired || out.NextAction != "" {
		t.Fatal("explained omissions still required a full read", out)
	}
	// A recorded state never hides a source that the current capture still has.
	current.SourceStates = []AssignmentSourceState{{ResourceKind: AssignmentContentComponent, SourceID: "current", State: AssignmentSourceHidden}}
	if out = compareAssignmentContextParts(current, current); len(out.Changes) != 0 || !out.Complete {
		t.Fatal("state for a present source changed the comparison", out)
	}
}
