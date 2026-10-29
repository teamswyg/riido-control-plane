package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTaskReadScopeMetadataPreservesAllFixedSelectedInspectorFields(t *testing.T) {
	var work taskReadScopeWork
	if err := json.Unmarshal([]byte(`{"id":"work","kind":"component","teamId":"team","priorities":[{"id":"priority-value","score":0}],"checklists":[{"id":"checklist-value","items":[]}],"attachments":[{"id":"attachment-value","fileSize":0}],"details":{"id":"work","title":"Title","isDone":false,"estimate":0,"createdBy":null},"github":[{"id":"pr-value","body":"stored body","comments":[]}],"figma":{"id":"figma-value","url":"https://figma.test/example"},"links":[{"id":"link-value","url":null}],"repositories":[{"id":"public-repo","fullName":"team/repo","isPrivate":false,"repositoryUrl":"https://github.test/team/repo","source":"project"}]}`), &work); err != nil {
		t.Fatal(err)
	}
	b, _ := newTaskReadScopeBuilder()
	part := b.captureWorkMetadata(AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "work"}, work)
	if !part.Complete || len(b.capture.Pages) != 1 {
		t.Fatalf("metadata capture=%+v", part)
	}
	for _, field := range []string{"priority-value", "checklist-value", "attachment-value", "pr-value", "figma-value", "link-value", "public-repo", `"isPrivate":false`, `"isDone":false`, `"estimate":0`, `"createdBy":null`} {
		if !strings.Contains(string(b.capture.Pages[0].Data), field) {
			t.Errorf("selected field missing: %s", field)
		}
	}
}
