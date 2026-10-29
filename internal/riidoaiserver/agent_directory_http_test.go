package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAgentDirectoryListsRequesterVisibleAgentsLive(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	code, body := f.read(AssignmentContextViewAgents, "", f.leased.LeaseToken)
	var resp struct {
		Live    bool               `json:"live"`
		Data    AgentDirectoryPage `json:"data"`
		Sources []AssignmentContextSource
	}
	if code != http.StatusOK || json.Unmarshal([]byte(body), &resp) != nil || !resp.Live || resp.Data.PolicyChecked || resp.Data.FetchedAt.IsZero() {
		t.Fatalf("status=%d body=%s", code, body)
	}
	got := map[string]AgentDirectoryEntry{}
	for _, entry := range resp.Data.Agents {
		got[entry.AgentID] = entry
		if entry.Callable != (len(entry.UnavailableReasons) == 0) || entry.RoleStatus != AgentDirectoryRoleDescribed {
			t.Fatalf("inconsistent entry: %+v", entry)
		}
	}
	if len(got) != 3 || resp.Data.Total != 3 || got["agent-private-cursor"].AgentID != "" || len(resp.Sources) != 3 {
		t.Fatalf("requester scope: %s", body)
	}
	if self := got["agent-owned-codex"]; !self.IsCurrentAgent || self.Callable || !strings.Contains(strings.Join(self.UnavailableReasons, ","), AgentDirectoryReasonSelf) {
		t.Fatalf("self entry: %+v", self)
	}
	for _, leaked := range []string{"instruction", "owner_principal_id", "user-1", "user-2", "반복 가능한 보조 작업"} {
		if strings.Contains(body, leaked) {
			t.Fatalf("leaked %q: %s", leaked, body)
		}
	}
	reads, _ := f.contexts.ListAssignmentContextReads(context.Background(), f.leased.ID)
	last := reads[len(reads)-1]
	if last.SourceKind != "agent" || last.SourcePart != "profile" || last.Via != AssignmentContextViaCLI || last.Status != AssignmentContextReadServed || last.SourceVersion == "" {
		t.Fatalf("read record: %+v", last)
	}
}

func TestAgentDirectoryShowRefusesHiddenAndUnknownAgents(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	code, body := f.read(AssignmentContextViewAgent, "agent-public-openclaw", f.leased.LeaseToken)
	var resp struct{ Data AgentDirectoryDetail }
	if code != http.StatusOK || json.Unmarshal([]byte(body), &resp) != nil || resp.Data.Agent.AgentID != "agent-public-openclaw" || resp.Data.Agent.IsCurrentAgent {
		t.Fatalf("show status=%d body=%s", code, body)
	}
	for _, target := range []string{"agent-private-cursor", "agent-missing", ""} {
		if code, body := f.read(AssignmentContextViewAgent, target, f.leased.LeaseToken); code != http.StatusForbidden || !strings.Contains(body, "outside") {
			t.Fatalf("target %q status=%d body=%s", target, code, body)
		}
	}
	reads, _ := f.contexts.ListAssignmentContextReads(context.Background(), f.leased.ID)
	if last := reads[len(reads)-1]; last.Status != AssignmentContextReadDenied {
		t.Fatalf("denied read not recorded: %+v", last)
	}
	if code, _ := f.read(AssignmentContextViewAgents, "", "wrong-lease"); code != http.StatusForbidden {
		t.Fatalf("lease check skipped: %d", code)
	}
}
