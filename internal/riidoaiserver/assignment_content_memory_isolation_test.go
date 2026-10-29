package riidoaiserver

import "testing"

func TestAssignmentContentReturnedManifestCannotMutateStoredScope(t *testing.T) {
	m := NewMemoryAssignmentContextStore()
	snapshot, pages := assignmentContentTestBundle()
	if err := m.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	got, _, err := m.LoadAssignmentContextSnapshot(t.Context(), "assignment")
	if err != nil {
		t.Fatal(err)
	}
	got.Content[0].SourceID = "outside"
	got.Content[0].Cursors[0] = "forged"
	again, _, err := m.LoadAssignmentContextSnapshot(t.Context(), "assignment")
	if err != nil || again.Content[0].SourceID != "work" || again.Content[0].Cursors[0] != "first" {
		t.Fatalf("returned manifest changed stored scope: %+v err=%v", again.Content, err)
	}
}
