package riidoaiserver

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommentHandoffInheritsContextAndServesChildCLI(t *testing.T) {
	contexts := NewMemoryAssignmentContextStore()
	f := newCommentCallFixtureWithContext(t, contexts)
	f.readHandoffCandidate(t, "agent-owned-codex")
	if w := f.complete(t, mentionHTML("agent-owned-codex")+"Compare the two plans."); w.Code != 200 {
		t.Fatalf("completion: %d %s", w.Code, w.Body.String())
	}
	child := f.threads(t)[1]
	snapshot, found, err := contexts.LoadAssignmentContextSnapshot(t.Context(), child.AssignmentID)
	if err != nil || !found || snapshot.TriggerKind != AssignmentContextTriggerAgentMention || snapshot.TriggerID != child.SourceMessageID || snapshot.Context.Document.Content != f.evidence.Context.Document.Content || snapshot.RequesterPrincipalID != f.evidence.PrincipalID {
		t.Fatalf("child did not inherit parent snapshot: %+v %t %v", snapshot, found, err)
	}
	f.client.deviceCredentials["device-dev-macbook"] = deviceCredentialRecord{deviceID: "device-dev-macbook", secretHash: sha256.Sum256([]byte(assignmentContextTestSecret)), ownerPrincipalID: "user-1", workspaceID: defaultAIAgentClientWorkspaceID}
	poll, err := f.assignments.PollAgent(t.Context(), "agent-owned-codex", PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
	if err != nil || poll.Assignment == nil || poll.Assignment.ID != child.AssignmentID {
		t.Fatalf("child lease: %+v %v", poll, err)
	}
	for _, view := range []string{AssignmentContextViewBase, AssignmentContextViewAgents, AssignmentContextViewAgent} {
		target := ""
		if view == AssignmentContextViewAgent {
			target = "agent-public-openclaw"
		}
		input, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: child.AssignmentID, LeaseToken: poll.Assignment.LeaseToken, View: view, Target: target, RequestID: "child-cli"})
		r := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/assignment-context", strings.NewReader(string(input)))
		r.Header.Set(deviceIDHeader, "device-dev-macbook")
		r.Header.Set(deviceSecretHeader, assignmentContextTestSecret)
		w := httptest.NewRecorder()
		f.s.Handler().ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("child CLI %s: %d %s", view, w.Code, w.Body.String())
		}
		if view == AssignmentContextViewBase && !strings.Contains(w.Body.String(), "Plan A costs 30") {
			t.Fatal("child CLI lost task context")
		}
		if view != AssignmentContextViewBase && (!strings.Contains(w.Body.String(), `"live":true`) || !strings.Contains(w.Body.String(), `"role_status"`)) {
			t.Fatal("child CLI did not return live role/status evidence")
		}
	}
	reads, err := contexts.ListAssignmentContextReads(t.Context(), child.AssignmentID)
	if err != nil || len(reads) == 0 {
		t.Fatalf("child CLI reads not recorded: %v", err)
	}
	for _, read := range reads {
		if read.Via != AssignmentContextViaCLI || read.AgentID != child.AgentID {
			t.Fatalf("parent prompt reads incorrectly attributed to child: %+v", read)
		}
	}
}
