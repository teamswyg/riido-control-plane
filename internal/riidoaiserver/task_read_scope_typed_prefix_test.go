package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestTaskReadScopeTypedPrefixRetainsCorrectAuditPart(t *testing.T) {
	for _, tc := range []struct{ part, want string }{{"document", "body"}, {"description", "body"}, {"document_alternate", "body"}, {"comments_page", "comments"}, {"metadata", "metadata"}, {"board_page", "metadata"}} {
		t.Run(tc.part, func(t *testing.T) {
			selector := &taskReadPartSelector{}
			first := typedJSONFixture("resource", tc.part, "version", `{"content":"unfinished prefix"}`, 0, selector)
			// Force an incomplete transport prefix even though it happens to be JSON.
			retainTypedPrefix(first, []byte(`{"content":"unfinished`))
			pages := readScopeTypedPrefixPages(AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "resource", TeamID: "team"}, first, 0)
			if len(pages) != 1 || len(pages[0].Sources) != 1 || pages[0].Sources[0].Part != tc.want {
				t.Fatalf("incorrect partial source: %+v", pages)
			}
			var got taskReadBodyPage
			if json.Unmarshal(pages[0].Data, &got) != nil || got.Complete == nil || *got.Complete || got.Content != first.Content || pages[0].Complete {
				t.Fatal("prefix became complete typed data")
			}
		})
	}
}
