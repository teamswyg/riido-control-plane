package riidoaiserver

import (
	"testing"
)

func inlineBoundaryFixture(t *testing.T, resources []AssignmentScopedResource, pages []AssignmentContentPage) *assignmentContextFixture {
	t.Helper()
	store := NewMemoryAssignmentContextStore()
	f := newAssignmentContextFixture(t, store)
	snapshot, found, err := store.LoadAssignmentContextSnapshot(t.Context(), f.leased.ID)
	if err != nil || !found {
		t.Fatal("missing synthetic initial snapshot", err)
	}
	snapshot.Content = nil
	snapshot.Context.TeamID = "team"
	for _, r := range resources {
		snapshot.Content = append(snapshot.Content, r.Parts...)
	}
	snapshot.Context.ReadScope = &AssignmentReadScope{SchemaVersion: taskReadScopeSchema, RootID: "task-private", TeamID: "team", Status: "available", Resources: resources}
	store.mu.Lock()
	delete(store.snapshots, f.leased.ID)
	store.mu.Unlock()
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	return f
}
