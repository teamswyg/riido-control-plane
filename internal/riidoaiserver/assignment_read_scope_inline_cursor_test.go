package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAssignmentScopedInlineRejectsLegacyCursor(t *testing.T) {
	resource, page := inlineBoundaryPart(AssignmentContentComponent, "task-private", AssignmentContentEditor, "first", "body", `{"content":"body"}`)
	f := inlineBoundaryFixture(t, []AssignmentScopedResource{resource}, []AssignmentContentPage{page})
	for _, view := range []string{AssignmentContextViewTask, AssignmentContextViewComments, AssignmentContextViewMeeting} {
		request, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: view, Target: "task-private", Cursor: "opaque-v2-must-use-content"})
		status, _ := f.do("/v1/agents/agent-owned-codex/assignment-context", string(request), f.device)
		if status != http.StatusBadRequest {
			t.Fatalf("inline view %s silently accepted/ignored cursor: %d", view, status)
		}
	}
}
