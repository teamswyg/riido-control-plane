package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestExplicitThreadReadsRetainedProjection(t *testing.T) {
	f := explicitReadFixture(t, func(f *assignmentContextFixture, v *explicitAssignmentResult, _ http.ResponseWriter, _ *http.Request) bool {
		v.ID = f.leased.TaskID
		return true
	})
	code, body := f.read("threads", f.leased.TaskID, f.leased.LeaseToken)
	if code != 200 || !strings.Contains(body, `"history_scope":"server_retained"`) || !strings.Contains(body, `"complete":false`) || strings.Contains(body, "lease_token") {
		t.Fatalf("status=%d body=%s", code, body)
	}
	req := AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: "thread", Target: f.leased.TaskID, ThreadID: "another-task-thread"}
	raw, _ := json.Marshal(req)
	code, body = f.do("/v1/agents/agent-owned-codex/assignment-context", string(raw), f.device)
	if code != 403 || strings.Contains(body, `"messages"`) {
		t.Fatalf("cross-task thread status=%d body=%s", code, body)
	}
}

func TestExplicitReadOptionsCannotSetAuthority(t *testing.T) {
	for _, key := range []string{"principal_id", "workspace_id", "root_id", "admin", "team-id", "url"} {
		if validExplicitOptions(AgentAssignmentContextRequest{View: "activity", Options: map[string]string{key: "value"}}) {
			t.Fatalf("accepted authority option %s", key)
		}
	}
	if validExplicitOptions(AgentAssignmentContextRequest{View: "base", Options: map[string]string{"page": "1"}}) {
		t.Fatal("base view changed to live")
	}
}
