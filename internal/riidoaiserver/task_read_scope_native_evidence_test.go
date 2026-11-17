package riidoaiserver

import "testing"

func TestTaskReadScopeNativeEvidenceStagesAndCentralParent(t *testing.T) {
	for _, kind := range []string{"component", "intake", "backlog"} {
		t.Run(kind, func(t *testing.T) {
			bundle := nativeTestCapture(t, kind)
			evidence := verifiedCommentDelivery{TaskID: "task-private", WorkspaceID: "workspace", PrincipalID: "requester", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task-private", ComponentType: "task"}, ReadScopeEvidence: bundle}}
			got, err := materializeCommentReadScope(t.Context(), evidence)
			if err != nil {
				t.Fatal(err)
			}
			if got.Component.ID != "task-private" || got.Component.ComponentType != "task" || got.Component.Kind != kind || got.Hierarchy.Project.ID != "project" || got.Hierarchy.ParentTask.ID != "" {
				t.Fatalf("wrong native context: %+v", got)
			}
			if got.ReadScope.WorkspaceID != "workspace" || (got.ReadScope.SchemaVersion != taskReadScopeSchema || got.ReadScope.SourceSchemaVersion != taskReadScopeNativeSchema) || got.ReadScope.Status != "available" {
				t.Fatalf("scope: %+v", got.ReadScope)
			}
			snapshot := AssignmentContextSnapshot{WorkspaceID: "workspace", Context: got, Content: got.readScopeCapture.Manifest}
			if err := validateAssignmentReadScopeManifest(snapshot); err != nil {
				t.Fatal(err)
			}
			for _, part := range snapshot.Content {
				if !part.Complete {
					t.Fatalf("incomplete native capture: %+v", part)
				}
			}
			evidence.WorkspaceID = "other"
			if _, err := materializeCommentReadScope(t.Context(), evidence); err == nil {
				t.Fatal("cross-workspace evidence accepted")
			}
			evidence.WorkspaceID = "workspace"
			bundle.BodyPages[0].Page.Version = "sha256:wrong"
			if _, err := materializeCommentReadScope(t.Context(), evidence); err == nil {
				t.Fatal("changed continuation accepted")
			}
		})
	}
}

func TestTaskReadScopeNativeRejectsScopeWorkspaceAndRootMismatch(t *testing.T) {
	bundle := nativeTestCapture(t, "component")
	for _, mutate := range []func(*taskReadScope){
		func(s *taskReadScope) { s.WorkspaceID = "" },
		func(s *taskReadScope) { s.Works[0].WorkspaceID = "other" },
		func(s *taskReadScope) { s.Works[0].TeamID = "old-team" },
	} {
		bundle = nativeTestCapture(t, "component")
		mutate(bundle.Scope)
		if err := validateTaskReadScope(*bundle.Scope, "task-private", "team"); err == nil {
			t.Fatal("invalid native scope accepted")
		}
	}
}
