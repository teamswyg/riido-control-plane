package riidoaiserver

import "testing"

func TestAssignmentReadScopeInheritancePublishesAllPartsAtomically(t *testing.T) {
	store := NewMemoryAssignmentContextStore()
	snapshot, pages := assignmentContentTestBundle()
	snapshot.RequesterPrincipalID = "requester"
	snapshot.WorkspaceID = "workspace"
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	server := Server{config: ServerConfig{AssignmentContext: store}}
	child := Assignment{ID: "child", TaskID: snapshot.TaskID, WorkspaceID: "workspace", CreatedBy: "requester", AgentID: "agent-b"}
	if err := server.InheritAssignmentContext(t.Context(), snapshot.AssignmentID, child, "mention"); err != nil {
		t.Fatal(err)
	}
	for i, want := range pages {
		got, found, err := store.LoadAssignmentContentPage(t.Context(), child.ID, i)
		if err != nil || !found || string(got.Data) != string(want.Data) {
			t.Fatalf("child content %d missing: %v %v", i, found, err)
		}
	}
	child.ID = "cross-workspace"
	child.WorkspaceID = "another"
	if err := server.InheritAssignmentContext(t.Context(), snapshot.AssignmentID, child, "mention"); err == nil {
		t.Fatal("cross workspace inherited content")
	}
}
