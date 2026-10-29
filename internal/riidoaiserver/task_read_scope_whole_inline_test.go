package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func assertReadScopeWholeInline(t *testing.T, f *assignmentContextFixture, resources []AssignmentScopedResource) {
	t.Helper()
	code, _ := f.read("base", "", f.leased.LeaseToken)
	if code != http.StatusOK {
		t.Fatalf("base=%d", code)
	}
	parts := 0
	for _, resource := range resources {
		view := "task"
		if resource.Kind == AssignmentContentMeeting {
			view = "meeting"
		}
		code, body := f.read(view, resource.ID, f.leased.LeaseToken)
		var response struct {
			Data assignmentScopedInlineView `json:"data"`
		}
		if code != http.StatusOK || json.Unmarshal([]byte(body), &response) != nil || len(response.Data.Parts) == 0 {
			t.Fatalf("inline %s=%d %s", resource.ID, code, body)
		}
		for _, part := range response.Data.Parts {
			if part.Page == nil || !part.Page.Complete || !part.Manifest.Complete || part.NextCursor != "" {
				t.Fatalf("normal part needs extra read: %s %+v", resource.ID, part)
			}
			parts++
		}
	}
	if parts != 120 {
		t.Fatalf("all normal metadata/body/description/comments or board parts=%d want120", parts)
	}
}
