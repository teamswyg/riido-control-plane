package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func contextChangesTestBundle(t *testing.T, snapshot AssignmentContextSnapshot, body string) (AssignmentContextSnapshot, []AssignmentContentPage) {
	t.Helper()
	b, err := newTaskReadScopeBuilder()
	if err != nil {
		t.Fatal(err)
	}
	resource := AssignmentScopedResource{Kind: AssignmentContentComponent, ID: snapshot.TaskID, TeamID: snapshot.Context.TeamID, Title: "Current title", Reasons: []taskReadScopeReason{{Kind: "root"}}}
	raw, _ := json.Marshal(map[string]string{"body": body})
	resource.Parts = []AssignmentContentManifest{b.addPart(resource, AssignmentContentEditor, "source-v1", []AssignmentContentPage{{Status: "available", Data: raw, Sources: []AssignmentContextSource{{Kind: "task", ID: snapshot.TaskID, Part: "body"}}}}, true)}
	snapshot.Context.Component.ID = snapshot.TaskID
	snapshot.Context.ReadScope = &AssignmentReadScope{SchemaVersion: taskReadScopeSchema, RootID: snapshot.TaskID, TeamID: snapshot.Context.TeamID, Status: "available", Resources: []AssignmentScopedResource{resource}}
	snapshot.Content = b.capture.Manifest
	return snapshot, b.capture.Pages
}

func contextChangesHTTPFixture(t *testing.T) (*assignmentContextFixture, AssignmentContextSnapshot) {
	t.Helper()
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	current, found, err := f.contexts.LoadAssignmentContextSnapshot(t.Context(), f.leased.ID)
	if err != nil || !found {
		t.Fatal("missing initial snapshot", err)
	}
	current.Context.TeamID = "fixture-team"
	current.ThreadID = "current-thread"
	previous := current
	previous.AssignmentID, previous.ThreadID = "previous", "previous-thread"
	previous.ConversationID = previous.ThreadID
	current.ConversationID = previous.ConversationID
	current.PreviousAssignmentID, current.PreviousThreadID = previous.AssignmentID, previous.ThreadID
	previous, pages := contextChangesTestBundle(t, previous, "OLD-PRIVATE-BODY")
	if err = f.contexts.SaveAssignmentContentBundle(t.Context(), previous, pages); err != nil {
		t.Fatal(err)
	}
	current, pages = contextChangesTestBundle(t, current, "NEW-PRIVATE-BODY")
	// Replace only this fixture's legacy capture with the equivalent typed bundle.
	f.contexts.mu.Lock()
	delete(f.contexts.snapshots, current.AssignmentID)
	f.contexts.mu.Unlock()
	if err = f.contexts.SaveAssignmentContentBundle(t.Context(), current, pages); err != nil {
		t.Fatal(err)
	}
	return f, previous
}

func contextChangesRequest(f *assignmentContextFixture, since, lease string) string {
	raw, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: lease, View: AssignmentContextViewChanges, Since: since, RequestID: "changes-read"})
	return string(raw)
}
