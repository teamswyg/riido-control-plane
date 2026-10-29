package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestAssignmentContentPacksTwentyFiveSourcesAndThreePartsWithoutDroppingPages(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(fmt.Sprint(durable), func(t *testing.T) {
			var store AssignmentContentStore = NewMemoryAssignmentContextStore()
			fixture := &contentBundleFixture{}
			if durable {
				store = newContentBundleDynamo(t, fixture)
			}
			snapshot := AssignmentContextSnapshot{AssignmentID: "packed", TaskID: "root"}
			var pages []AssignmentContentPage
			for i := range 25 {
				for _, part := range []AssignmentContentPart{AssignmentContentEditor, AssignmentContentDescription, AssignmentContentComments} {
					id, cursor := fmt.Sprintf("work-%d", i), fmt.Sprintf("cursor-%d-%s", i, part)
					snapshot.Content = append(snapshot.Content, AssignmentContentManifest{ResourceKind: AssignmentContentComponent, SourceID: id, Part: part, Version: "v1", Status: "available", Complete: true, Cursors: []string{cursor}})
					pages = append(pages, AssignmentContentPage{ResourceKind: AssignmentContentComponent, Cursor: cursor, SourceID: id, Part: part, Version: "v1", Status: "available", Complete: true, Data: json.RawMessage(`{"content":"small full body"}`), Sources: []AssignmentContextSource{{Kind: "related_task", ID: id, Part: "body", Version: "v1"}}})
				}
			}
			if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
				t.Fatal(err)
			}
			for i, want := range pages {
				got, found, err := store.LoadAssignmentContentPage(t.Context(), "packed", i)
				if err != nil || !found || got.SourceID != want.SourceID || got.Cursor != want.Cursor || !got.Complete {
					t.Fatalf("logical page %d=%+v found=%v err=%v", i, got, found, err)
				}
			}
			if durable && len(fixture.items) != 2 {
				t.Fatalf("small logical pages did not share one physical block: %d items", len(fixture.items))
			}
		})
	}
}
