package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestContextChangesRequestRejectsUnboundedOrMixedSelection(t *testing.T) {
	f, _ := contextChangesHTTPFixture(t)
	base := AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: AssignmentContextViewChanges, Since: "previous"}
	for _, mutate := range []func(*AgentAssignmentContextRequest){
		func(r *AgentAssignmentContextRequest) { r.Since = "" },
		func(r *AgentAssignmentContextRequest) { r.Since = strings.Repeat("a", 257) },
		func(r *AgentAssignmentContextRequest) { r.Target = "other" },
		func(r *AgentAssignmentContextRequest) { r.Cursor = "cursor" },
		func(r *AgentAssignmentContextRequest) { r.MaxBytes = 8192 },
		func(r *AgentAssignmentContextRequest) { r.View = AssignmentContextViewBase },
	} {
		req := base
		mutate(&req)
		raw, _ := json.Marshal(req)
		if code, _ := f.do(contextChangesTestPath, string(raw), f.device); code != http.StatusBadRequest {
			t.Fatal("invalid selection accepted", code)
		}
	}
}
