package riidoaiserver

import "testing"

func TestAssignmentReadScopeViewerUsesCurrentLeavesAndAllCurrentPartVersions(t *testing.T) {
	resource := AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "selected", Title: "Current selected", Reasons: []taskReadScopeReason{{Kind: "mention"}}}
	current := AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "root"}, ReadScope: &AssignmentReadScope{RootID: "root", Resources: []AssignmentScopedResource{resource}}, readScopeCapture: &assignmentReadScopeCapture{Manifest: []AssignmentContentManifest{{Complete: true, Cursors: []string{"editor", "description", "comment"}}}, Pages: []AssignmentContentPage{
		{Cursor: "editor", Sources: []AssignmentContextSource{{Kind: "related_task", ID: "selected", Part: "body", Version: "editor-version"}}},
		{Cursor: "description", Sources: []AssignmentContextSource{{Kind: "related_task", ID: "selected", Part: "body", Version: "description-version"}}},
		{Cursor: "comment", Sources: []AssignmentContextSource{{Kind: "comment", ID: "visible-comment", Part: "body", Version: "comment-version"}}},
	}}}
	visible := visibleRunReadSources(current)
	for _, key := range []runReadSourceKey{{"related_task", "selected", "body"}, {"related_task", "selected", "title"}, {"comment", "visible-comment", "body"}} {
		if !visible[key] {
			t.Errorf("current source hidden: %+v", key)
		}
	}
	if visible[runReadSourceKey{"related_task", "removed", "body"}] {
		t.Fatal("retained source leaked into current access")
	}
	metadata := currentRunReadMetadata(current)
	for _, version := range []string{"editor-version", "description-version"} {
		changed := runReadModifiedSinceRead(AssignmentContextRead{Status: "served", SourceVersion: version}, metadata[runReadSourceKey{"related_task", "selected", "body"}])
		if changed == nil || *changed {
			t.Errorf("current part version not recognized: %s", version)
		}
	}
}
