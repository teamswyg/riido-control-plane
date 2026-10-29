package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func inlineBoundaryContent(t *testing.T, f *assignmentContextFixture, target, cursor string) (int, string) {
	t.Helper()
	raw, err := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: AssignmentContextViewContent, Target: target, Cursor: cursor})
	if err != nil {
		t.Fatal(err)
	}
	return f.do("/v1/agents/agent-owned-codex/assignment-context", string(raw), f.device)
}
