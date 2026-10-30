package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestAssignmentContentDynamoVisitsPagesAcrossBlocksInOrder(t *testing.T) {
	fixture := &contentBundleFixture{}
	store := newContentBundleDynamo(t, fixture)
	manifest := AssignmentContentManifest{ResourceKind: AssignmentContentComponent, SourceID: "work", Part: AssignmentContentDescription, Version: "v1", Status: "available", Complete: true}
	var pages []AssignmentContentPage
	for i := range 7 {
		cursor := fmt.Sprintf("page-%d", i)
		manifest.Cursors = append(manifest.Cursors, cursor)
		data, _ := json.Marshal(map[string]string{"content": strings.Repeat(string(rune('a'+i)), 20_000)})
		pages = append(pages, AssignmentContentPage{ResourceKind: AssignmentContentComponent, Cursor: cursor, SourceID: "work", Part: AssignmentContentDescription, Version: "v1", Offset: i * 20_000, Status: "partial", Complete: i == 6, Data: data, Sources: []AssignmentContextSource{{Kind: "task", ID: "work", Part: "body", Version: "v1"}}})
	}
	pages[6].Status = "available"
	snapshot := AssignmentContextSnapshot{AssignmentID: "assignment", TaskID: "work", Content: []AssignmentContentManifest{manifest}}
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	saved, _, err := store.LoadAssignmentContextSnapshot(t.Context(), "assignment")
	if err != nil || len(saved.ContentBlocks) < 3 {
		t.Fatalf("expected several physical blocks, got %d err=%v", len(saved.ContentBlocks), err)
	}
	var visited []string
	if err := store.VisitAssignmentContentPages(t.Context(), "assignment", 1, 6, func(page AssignmentContentPage) bool {
		visited = append(visited, page.Cursor)
		return len(visited) < 5
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(visited, ",") != "page-1,page-2,page-3,page-4,page-5" {
		t.Fatalf("visited=%v", visited)
	}
	single, found, err := store.LoadAssignmentContentPage(t.Context(), "assignment", 6)
	if err != nil || !found || single.Cursor != "page-6" || !single.Complete {
		t.Fatalf("single=%+v found=%v err=%v", single, found, err)
	}
}
