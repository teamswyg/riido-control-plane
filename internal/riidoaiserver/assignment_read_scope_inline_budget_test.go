package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAssignmentScopedInlineOversizeRetainsOriginalFirstCursorWithoutBodyAudit(t *testing.T) {
	encoded, _ := json.Marshal(map[string]string{"content": strings.Repeat("x", 50<<10)})
	resource, page := inlineBoundaryPart(AssignmentContentComponent, "task-private", AssignmentContentEditor, "large-first", "body", string(encoded))
	resource.Parts[0].Cursors = append(resource.Parts[0].Cursors, "large-next")
	page.Complete = false
	next := page
	next.Cursor = "large-next"
	next.Offset = 50 << 10
	next.Complete = true
	next.Data = json.RawMessage(`{"content":"tail"}`)
	f := inlineBoundaryFixture(t, []AssignmentScopedResource{resource}, []AssignmentContentPage{page, next})
	code, body := f.read("task", "task-private", f.leased.LeaseToken)
	var response struct {
		Data assignmentScopedInlineView `json:"data"`
	}
	if code != http.StatusOK || json.Unmarshal([]byte(body), &response) != nil || len(response.Data.Parts) != 1 {
		t.Fatalf("oversize inline response=%d %s", code, body)
	}
	part := response.Data.Parts[0]
	raw, _ := json.Marshal(response.Data)
	if len(raw) > assignmentScopedInlineBytes || part.Page != nil || len(part.Manifest.Cursors) != 2 || part.Manifest.Cursors[0] != "large-first" || part.Manifest.NextAction != "riido context part task-private --cursor large-first" || part.NextCursor != "" {
		t.Fatal("large part was truncated or its first captured cursor was lost")
	}
	reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range reads {
		if r.Via == AssignmentContextViaCLI && r.Status == AssignmentContextReadServed && r.SourcePart == "body" {
			t.Fatal("omitted body was audited as read")
		}
	}
	code, body = inlineBoundaryContent(t, f, "task-private", "large-first")
	if code != http.StatusOK || !strings.Contains(body, `"next_cursor":"large-next"`) {
		t.Fatalf("explicit large captured page unavailable: %d", code)
	}
}
