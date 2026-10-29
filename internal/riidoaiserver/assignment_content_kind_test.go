package riidoaiserver

import "testing"

func TestAssignmentContentSameIDAcrossKindsKeepsSeparateParts(t *testing.T) {
	snapshot, pages := assignmentContentTestBundle()
	first, second := snapshot.Content[0], snapshot.Content[0]
	first.Cursors, second.Cursors = []string{"first"}, []string{"last"}
	second.ResourceKind = AssignmentContentBacklog
	snapshot.Content = []AssignmentContentManifest{first, second}
	pages[0].Complete, pages[1].Complete = true, true
	pages[1].ResourceKind = AssignmentContentBacklog
	pages[1].Sources[0].Kind = "related_backlog"
	store := NewMemoryAssignmentContextStore()
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	for i, kind := range []AssignmentContentKind{AssignmentContentComponent, AssignmentContentBacklog} {
		got, found, err := store.LoadAssignmentContentPage(t.Context(), "assignment", i)
		if err != nil || !found || got.ResourceKind != kind {
			t.Fatalf("kind %s returned %+v found=%v err=%v", kind, got, found, err)
		}
	}
}

func TestAssignmentContentPageKindMustMatchManifest(t *testing.T) {
	snapshot, pages := assignmentContentTestBundle()
	pages[0].ResourceKind = AssignmentContentIntake
	if err := NewMemoryAssignmentContextStore().SaveAssignmentContentBundle(t.Context(), snapshot, pages); err == nil {
		t.Fatal("page table kind escaped manifest identity")
	}
}
