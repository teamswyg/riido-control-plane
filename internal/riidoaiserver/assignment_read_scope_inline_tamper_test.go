package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentScopedInlineRejectsResourceManifestMismatch(t *testing.T) {
	for _, field := range []string{"kind", "source", "part", "version", "cursor"} {
		t.Run(field, func(t *testing.T) {
			resource, page := inlineBoundaryPart(AssignmentContentComponent, "task-private", AssignmentContentEditor, "editor-first", "body", `{"content":"must-not-leak"}`)
			f := inlineBoundaryFixture(t, []AssignmentScopedResource{resource}, []AssignmentContentPage{page})
			// Simulate stored metadata corruption after a valid atomic publication.
			f.contexts.mu.Lock()
			snapshot := f.contexts.snapshots[f.leased.ID]
			part := &snapshot.Context.ReadScope.Resources[0].Parts[0]
			switch field {
			case "kind":
				part.ResourceKind = AssignmentContentBacklog
			case "source":
				part.SourceID = "other"
			case "part":
				part.Part = AssignmentContentDescription
			case "version":
				part.Version = "changed"
			case "cursor":
				part.Cursors = []string{"invented"}
			}
			f.contexts.snapshots[f.leased.ID] = snapshot
			f.contexts.mu.Unlock()
			status, body := f.read("task", "task-private", f.leased.LeaseToken)
			if status < 400 || status > 599 || strings.Contains(body, "must-not-leak") {
				t.Fatalf("mismatched %s metadata was served: %d %s", field, status, body)
			}
			reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range reads {
				if r.Via == AssignmentContextViaCLI && r.Status == AssignmentContextReadServed {
					t.Fatal("failed manifest read recorded served source")
				}
			}
		})
	}
}
