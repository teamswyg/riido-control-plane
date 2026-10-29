package riidoaiserver

import "testing"

func TestAssignmentReadScopeRetainedViewerRejectsMismatchedSnapshotIdentity(t *testing.T) {
	for name, mutate := range map[string]func(*AssignmentContextSnapshot){
		"assignment": func(s *AssignmentContextSnapshot) { s.AssignmentID = "other" },
		"workspace":  func(s *AssignmentContextSnapshot) { s.WorkspaceID = "other" },
		"task":       func(s *AssignmentContextSnapshot) { s.TaskID = "other" },
		"root":       func(s *AssignmentContextSnapshot) { s.Context.ReadScope.RootID = "other" },
		"schema":     func(s *AssignmentContextSnapshot) { s.Context.ReadScope.SchemaVersion = "unknown" },
	} {
		t.Run(name, func(t *testing.T) {
			s, _, _ := runReadsTestServer(t, nil)
			reader := installCapturedScopeViewer(&s)
			store := s.config.AssignmentContext.(*MemoryAssignmentContextStore)
			store.mu.Lock()
			snapshot := store.snapshots["assignment"]
			mutate(&snapshot)
			store.snapshots["assignment"] = snapshot
			store.mu.Unlock()
			response := runReadsRequest(s, "/v1/client/ai-agent/runs/run/reads")
			if response.Code != 503 || len(reader.requests) != 0 {
				t.Fatal("mismatched stored identity selected a live viewer read")
			}
		})
	}
}
