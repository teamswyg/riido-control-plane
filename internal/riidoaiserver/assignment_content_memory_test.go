package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestAssignmentContentBundlePublishesCompleteCapturedPages(t *testing.T) {
	store := NewMemoryAssignmentContextStore()
	snapshot, pages := assignmentContentTestBundle()
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	got, found, err := store.LoadAssignmentContextSnapshot(t.Context(), "assignment")
	if err != nil || !found || len(got.Content) != 1 || !got.Content[0].Complete {
		t.Fatalf("snapshot=%+v found=%v err=%v", got, found, err)
	}
	for i, want := range pages {
		page, found, err := store.LoadAssignmentContentPage(t.Context(), "assignment", i)
		if err != nil || !found || string(page.Data) != string(want.Data) || page.Cursor != want.Cursor {
			t.Fatalf("page %d=%+v found=%v err=%v", i, page, found, err)
		}
	}
	if _, found, _ := store.LoadAssignmentContentPage(t.Context(), "another", 0); found {
		t.Fatal("page escaped assignment")
	}
}

func TestAssignmentContentBundleRejectsUnpublishedPageAndSharedBudget(t *testing.T) {
	for _, mutation := range []string{"orphan", "budget", "version"} {
		t.Run(mutation, func(t *testing.T) {
			store := NewMemoryAssignmentContextStore()
			snapshot, pages := assignmentContentTestBundle()
			switch mutation {
			case "orphan":
				snapshot.Content[0].Cursors = snapshot.Content[0].Cursors[:1]
			case "budget":
				snapshot.CommentPages = 9
			case "version":
				pages[1].Version = "changed"
			}
			if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err == nil {
				t.Fatal("invalid bundle accepted")
			}
			if _, found, _ := store.LoadAssignmentContextSnapshot(t.Context(), "assignment"); found {
				t.Fatal("failed bundle published a manifest")
			}
		})
	}
}

func assignmentContentTestBundle() (AssignmentContextSnapshot, []AssignmentContentPage) {
	snapshot := AssignmentContextSnapshot{AssignmentID: "assignment", TaskID: "work", Content: []AssignmentContentManifest{{ResourceKind: AssignmentContentComponent, SourceID: "work", Part: AssignmentContentEditor, Version: "v1", Status: "available", Complete: true, Cursors: []string{"first", "last"}}}}
	var pages []AssignmentContentPage
	for i, cursor := range snapshot.Content[0].Cursors {
		pages = append(pages, AssignmentContentPage{ResourceKind: AssignmentContentComponent, Cursor: cursor, SourceID: "work", Part: AssignmentContentEditor, Version: "v1", Offset: i * 2, Status: "available", Complete: i == 1, Data: json.RawMessage(`{"content":"ab"}`), Sources: []AssignmentContextSource{{Kind: "task", ID: "work", Part: "body", Version: "v1"}}})
	}
	return snapshot, pages
}
