package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentScopedInlineRejectsDuplicateGlobalManifestKey(t *testing.T) {
	for _, view := range []string{"task", "content"} {
		t.Run(view, func(t *testing.T) {
			resource, page := inlineBoundaryPart(AssignmentContentComponent, "task-private", AssignmentContentEditor, "editor-first", "body", `{"content":"must-not-leak"}`)
			f := inlineBoundaryFixture(t, []AssignmentScopedResource{resource}, []AssignmentContentPage{page})
			f.contexts.mu.Lock()
			snapshot := f.contexts.snapshots[f.leased.ID]
			duplicate := snapshot.Content[0]
			duplicate.Version = "unverified-version"
			duplicate.Cursors = []string{}
			snapshot.Content = append([]AssignmentContentManifest{duplicate}, snapshot.Content...)
			f.contexts.snapshots[f.leased.ID] = snapshot
			f.contexts.mu.Unlock()
			var status int
			var body string
			if view == "content" {
				status, body = inlineBoundaryContent(t, f, "task-private", "editor-first")
			} else {
				status, body = f.read("task", "task-private", f.leased.LeaseToken)
			}
			if status < 400 || strings.Contains(body, "must-not-leak") {
				t.Fatalf("duplicate manifest served: %d %s", status, body)
			}
			reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
			if err != nil {
				t.Fatal(err)
			}
			for _, read := range reads {
				if read.Via == AssignmentContextViaCLI && read.Status == AssignmentContextReadServed {
					t.Fatal("corrupt manifest recorded served audit")
				}
			}
		})
	}
}
