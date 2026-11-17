package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestAssignmentReadScopeNativePreservesDaemonV1ManifestBudget(t *testing.T) {
	bundle := nativeTestCapture(t, "component")
	got, err := materializeCommentReadScope(t.Context(), verifiedCommentDelivery{TaskID: "task-private", WorkspaceID: "workspace", PrincipalID: "human", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task-private"}, ReadScopeEvidence: bundle}})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(got)
	// The existing daemon contextbroker detects this exact nested field before
	// choosing the typed (paged) budget. Additional fields remain opaque JSON.
	var oldDaemon struct {
		ReadScope *struct {
			SchemaVersion string `json:"schema_version"`
		} `json:"read_scope"`
	}
	if json.Unmarshal(raw, &oldDaemon) != nil || oldDaemon.ReadScope == nil || oldDaemon.ReadScope.SchemaVersion != "riido-read-scope.v1" {
		t.Fatal("native manifest no longer selects existing daemon typed budget")
	}
	if got.ReadScope.SourceSchemaVersion != taskReadScopeNativeSchema || got.ReadScope.WorkspaceID != "workspace" {
		t.Fatal("producer identity lost")
	}
	snapshot := AssignmentContextSnapshot{AssignmentID: "assignment", WorkspaceID: "workspace", TaskID: "task-private", Context: got, Content: got.readScopeCapture.Manifest}
	if err := validateAssignmentReadScopeManifest(snapshot); err != nil {
		t.Fatal(err)
	}
	snapshot.Context.ReadScope.Resources[0].WorkspaceID = "other"
	if err := validateAssignmentReadScopeManifest(snapshot); err == nil {
		t.Fatal("foreign native resource accepted")
	}
}

func TestAssignmentReadScopeNativeRetainedViewerRechecksCurrentPermission(t *testing.T) {
	s, _, _ := runReadsTestServer(t, nil)
	reader := installCapturedScopeViewer(&s)
	store := s.config.AssignmentContext.(*MemoryAssignmentContextStore)
	store.mu.Lock()
	snapshot := store.snapshots["assignment"]
	snapshot.Context.ReadScope.SourceSchemaVersion = taskReadScopeNativeSchema
	snapshot.Context.ReadScope.WorkspaceID = defaultAIAgentClientWorkspaceID
	store.snapshots["assignment"] = snapshot
	store.mu.Unlock()
	response := runReadsRequest(s, "/v1/client/ai-agent/runs/run/reads")
	if response.Code != 200 || len(reader.requests) != 1 || !reader.requests[0].capturedReadScope || reader.requests[0].BearerToken != "viewer-token" {
		t.Fatal("native retained viewer did not request fresh viewer authority")
	}
	if _, err := s.currentRunReadRequiresScope(t.Context(), "assignment", "other", "task"); err == nil {
		t.Fatal("retained native workspace mismatch accepted")
	}
}
