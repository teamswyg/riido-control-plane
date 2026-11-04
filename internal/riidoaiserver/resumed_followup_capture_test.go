package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestResumedFollowupKeepsCurrentSnapshotWithoutFalsePromptReads(t *testing.T) {
	for _, scoped := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "scoped"}[scoped], func(t *testing.T) {
			f := newNativeResumeHTTPFixture(t)
			contexts := NewMemoryAssignmentContextStore()
			auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{PrincipalID: "user-1", Token: "native-owner-token", Scopes: []string{"riido:*"}}})
			if err != nil {
				t.Fatal(err)
			}
			c := aiAgentTaskContextHTTPFixture()
			c.Component.ID, c.TeamID = "task-new", "team"
			if scoped {
				typed, pages := contextChangesTestBundle(t, AssignmentContextSnapshot{TaskID: c.Component.ID, Context: c}, "captured task body")
				c = typed.Context
				c.readScopeCapture = &assignmentReadScopeCapture{Manifest: typed.Content, Pages: pages}
			}
			reader := &assignmentHTTPTaskContextReader{contextSnapshot: c}
			f.handler = NewServer(ServerConfig{Assignment: f.store, AIAgentClient: f.client, TaskContext: reader, Authorizer: auth, AssignmentContext: contexts, AssignmentReadScope: scoped}).Handler()
			first := f.finishFirst(t, AssignmentCompleted)
			initialReads, err := contexts.ListAssignmentContextReads(t.Context(), first.AssignmentID)
			if err != nil || len(initialReads) == 0 {
				t.Fatal("initial full prompt lost read evidence", err)
			}
			reader.contextSnapshot.Document.Content = "updated task document retained for CLI"
			var next AIAgentTaskActionResponse
			f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "Continue with the new requirements."}, http.StatusAccepted, &next)
			snapshot, found, err := contexts.LoadAssignmentContextSnapshot(t.Context(), next.AssignmentID)
			if err != nil || !found || snapshot.Context.Component.ID != "task-new" || snapshot.Context.Document.Content != "updated task document retained for CLI" {
				t.Fatal("resumed followup lost current CLI snapshot", err)
			}
			reads, err := contexts.ListAssignmentContextReads(t.Context(), next.AssignmentID)
			if err != nil || len(reads) != 0 {
				t.Fatalf("omitted snapshot marked as read: %+v err=%v", reads, err)
			}
			projection, found, err := f.store.LoadAssignmentProjection(t.Context(), next.AssignmentID)
			if err != nil || !found || strings.Contains(projection.Assignment.Prompt, snapshot.Context.Document.Content) {
				t.Fatal("stored snapshot was repeated in resumed prompt", err)
			}
		})
	}
}
