package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAssignmentScopedInlineSmallFirstPartsAuditOnlyDeliveredPages(t *testing.T) {
	metadata, metaPage := inlineBoundaryPart(AssignmentContentComponent, "task-private", AssignmentContentMetadata, "metadata-first", "metadata", `{"title":"Stored metadata"}`)
	editor, editorPage := inlineBoundaryPart(AssignmentContentComponent, "task-private", AssignmentContentEditor, "editor-first", "body", `{"content":"first"}`)
	editor.Parts[0].Cursors = append(editor.Parts[0].Cursors, "editor-next")
	editorPage.Complete = false
	next := editorPage
	next.Cursor = "editor-next"
	next.Offset = 5
	next.Complete = true
	next.Data = json.RawMessage(`{"content":"second"}`)
	metadata.Parts = append(metadata.Parts, editor.Parts...)
	f := inlineBoundaryFixture(t, []AssignmentScopedResource{metadata}, []AssignmentContentPage{metaPage, editorPage, next})
	code, body := f.read("task", "task-private", f.leased.LeaseToken)
	var response struct {
		Data assignmentScopedInlineView `json:"data"`
	}
	if code != http.StatusOK || json.Unmarshal([]byte(body), &response) != nil || len(response.Data.Parts) != 2 {
		t.Fatalf("inline response=%d %s", code, body)
	}
	raw, _ := json.Marshal(response.Data)
	if len(raw) > assignmentScopedInlineBytes || response.Data.Parts[0].Page == nil || response.Data.Parts[1].Page == nil || response.Data.Parts[1].NextCursor != "editor-next" || strings.Contains(body, `"second"`) {
		t.Fatal("first parts bound/continuation changed")
	}
	if response.Data.Parts[1].Manifest.NextAction != "riido context part task-private --cursor editor-next" {
		t.Fatal("first-page follow-up lost captured next cursor")
	}
	reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, r := range reads {
		if r.Via == AssignmentContextViaCLI && r.Status == AssignmentContextReadServed {
			counts[r.SourcePart+":"+r.SourceVersion]++
		}
	}
	if counts["metadata:v1"] != 1 || counts["body:v1"] != 1 {
		t.Fatalf("delivered first-page audits=%v", counts)
	}
}
