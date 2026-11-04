package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestContextChangesMetadataDoesNotDiscardExistingReadableSnapshot(t *testing.T) {
	snapshot, pages := contextChangesTestBundle(t, contextChangesTestSnapshot("near-limit"), "captured body")
	part := snapshot.Content[0]
	snapshot.Content[0].Digest = ""
	snapshot.Context.ReadScope.Resources[0].Parts[0].Digest = ""
	clearAssignmentContextLineage(&snapshot)
	_, refs, err := packAssignmentContentPages(pages)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.ContentBlocks = refs
	raw, _ := json.Marshal(snapshot)
	snapshot.Context.Document.Content = strings.Repeat("x", contextSnapshotItemBytes-128-len(raw)-8)
	if err := validateAssignmentContentBundle(snapshot, pages); err != nil {
		t.Fatal("baseline must be readable", err)
	}
	snapshot.Content[0].Digest = part.Digest
	snapshot.Context.ReadScope.Resources[0].Parts[0].Digest = part.Digest
	snapshot.ConversationID, snapshot.ThreadID = "conversation", "thread"
	if err := validateAssignmentContentBundle(snapshot, pages); err == nil {
		t.Fatal("fixture did not exceed metadata budget")
	}
	c := snapshot.Context
	c.readScopeCapture = &assignmentReadScopeCapture{Manifest: snapshot.Content, Pages: pages}
	store := NewMemoryAssignmentContextStore()
	server := NewServer(ServerConfig{AssignmentContext: store, AssignmentReadScope: true})
	server.captureAssignmentReadScope(t.Context(), snapshot, &c, true)
	got, found, err := store.LoadAssignmentContextSnapshot(t.Context(), snapshot.AssignmentID)
	if err != nil || !found {
		t.Fatal("comparison metadata discarded entire readable capture", err)
	}
	if got.ConversationID != "" || got.Content[0].Digest != "" || got.Context.ReadScope.Resources[0].Parts[0].Digest != "" {
		t.Fatal("fallback still claims usable comparison")
	}
	if c.ReadScope.Resources[0].Parts[0].Digest == "" || c.readScopeCapture.Manifest[0].Digest == "" {
		t.Fatal("fallback mutated input capture")
	}
	page, found, err := store.LoadAssignmentContentPage(t.Context(), snapshot.AssignmentID, 0)
	if err != nil || !found || string(page.Data) != string(pages[0].Data) {
		t.Fatal("fallback lost captured body", err)
	}
}
