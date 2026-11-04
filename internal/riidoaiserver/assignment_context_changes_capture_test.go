package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestContextChangesUnavailableBaselineUsesRecovery(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	body, _ := json.Marshal(map[string]string{"assignment_id": f.leased.ID, "lease_token": f.leased.LeaseToken, "view": "changes", "since": "unavailable-baseline", "request_id": "cli-changes"})
	code, out := f.do("/v1/agents/agent-owned-codex/assignment-context", string(body), f.device)
	if code != http.StatusGone {
		t.Fatalf("baseline recovery status=%d, want410; body=%s", code, out)
	}
	reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
	if err != nil || reads[len(reads)-1].SourcePart != "metadata" {
		t.Fatal("refusal used unsupported public audit part", err)
	}
}

func TestContextSnapshotCapturesServerSelectedConversation(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	contexts := NewMemoryAssignmentContextStore()
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{PrincipalID: "user-1", Token: "native-owner-token", Scopes: []string{"riido:*"}}})
	if err != nil {
		t.Fatal(err)
	}
	c := aiAgentTaskContextHTTPFixture()
	c.Component.ID = "task-new"
	c.TeamID = "team"
	typed, pages := contextChangesTestBundle(t, AssignmentContextSnapshot{TaskID: c.Component.ID, Context: c}, "typed body")
	c = typed.Context
	c.readScopeCapture = &assignmentReadScopeCapture{Manifest: typed.Content, Pages: pages}
	f.handler = NewServer(ServerConfig{Assignment: f.store, AIAgentClient: f.client, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: c}, Authorizer: auth, AssignmentContext: contexts, AssignmentReadScope: true}).Handler()
	first := f.finishFirst(t, AssignmentCompleted)
	var next AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "이어서 해줘"}, http.StatusAccepted, &next)
	snapshot, found, err := contexts.LoadAssignmentContextSnapshot(context.Background(), next.AssignmentID)
	if err != nil || !found {
		t.Fatal("missing followup capture")
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]json.RawMessage
	if err = json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"conversation_id": first.ThreadID, "thread_id": next.ThreadID, "previous_assignment_id": first.AssignmentID}
	for key, value := range want {
		var got string
		if json.Unmarshal(wire[key], &got) != nil || got != value {
			t.Fatalf("server-selected %s missing or different", key)
		}
	}
}

func TestContextLegacyCaptureDoesNotSpendBudgetOnComparisonMetadata(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	snapshot, found, err := f.contexts.LoadAssignmentContextSnapshot(t.Context(), f.leased.ID)
	if err != nil || !found {
		t.Fatal("legacy capture missing", err)
	}
	if snapshot.ConversationID != "" || snapshot.ThreadID != "" || snapshot.PreviousAssignmentID != "" || snapshot.PreviousThreadID != "" {
		t.Fatal("legacy capture stores unusable comparison metadata")
	}
}
