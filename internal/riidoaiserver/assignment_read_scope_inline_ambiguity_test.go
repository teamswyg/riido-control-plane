package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAssignmentScopedInlineAmbiguousWorkIDNeedsExactKindCursor(t *testing.T) {
	component, first := inlineBoundaryPart(AssignmentContentComponent, "task-private", AssignmentContentEditor, "component-body", "body", `{"content":"component-only"}`)
	backlog, second := inlineBoundaryPart(AssignmentContentBacklog, "task-private", AssignmentContentEditor, "backlog-body", "body", `{"content":"backlog-only"}`)
	f := inlineBoundaryFixture(t, []AssignmentScopedResource{component, backlog}, []AssignmentContentPage{first, second})
	status, _ := f.read("task", "task-private", f.leased.LeaseToken)
	if status != http.StatusForbidden {
		t.Fatalf("ambiguous work ID=%d want403", status)
	}
	for _, tc := range []struct {
		cursor  string
		kind    AssignmentContentKind
		content string
	}{{"component-body", AssignmentContentComponent, "component-only"}, {"backlog-body", AssignmentContentBacklog, "backlog-only"}} {
		code, body := inlineBoundaryContent(t, f, "task-private", tc.cursor)
		var response struct {
			Data AssignmentContentPage `json:"data"`
		}
		if code != http.StatusOK || json.Unmarshal([]byte(body), &response) != nil || response.Data.ResourceKind != tc.kind || !strings.Contains(string(response.Data.Data), tc.content) {
			t.Fatalf("opaque cursor lost kind identity: %d %s", code, body)
		}
	}
}
