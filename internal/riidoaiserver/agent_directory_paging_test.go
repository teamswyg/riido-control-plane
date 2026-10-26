package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAgentDirectoryWorksWithoutSnapshotAndPagesPastEnd(t *testing.T) {
	contexts := NewMemoryAssignmentContextStore()
	f := newAssignmentContextFixture(t, contexts)
	delete(contexts.snapshots, f.leased.ID)
	body, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: AssignmentContextViewAgents, Cursor: "1"})
	code, out := f.do("/v1/agents/agent-owned-codex/assignment-context", string(body), f.device)
	var resp struct{ Data AgentDirectoryPage }
	if code != http.StatusOK || json.Unmarshal([]byte(out), &resp) != nil || len(resp.Data.Agents) != 0 || resp.Data.Total != 3 || resp.Data.NextCursor != "" || resp.Data.Cursor != "1" {
		t.Fatalf("status=%d body=%s", code, out)
	}
}
