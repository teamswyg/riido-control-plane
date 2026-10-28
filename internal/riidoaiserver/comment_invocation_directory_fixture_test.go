package riidoaiserver

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func (f *commentCallFixture) readHandoffCandidate(t *testing.T, target string) {
	t.Helper()
	f.client.deviceCredentials["device-shared-studio"] = deviceCredentialRecord{deviceID: "device-shared-studio", secretHash: sha256.Sum256([]byte(assignmentContextTestSecret)), ownerPrincipalID: "user-2", workspaceID: defaultAIAgentClientWorkspaceID}
	poll, err := f.assignments.PollAgent(t.Context(), f.source.AgentID, PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
	if err != nil || poll.Assignment == nil || poll.Assignment.ID != f.source.AssignmentID {
		t.Fatalf("parent lease: %+v %v", poll, err)
	}
	for _, view := range []string{AssignmentContextViewAgents, AssignmentContextViewAgent} {
		queryTarget := ""
		if view == AssignmentContextViewAgent {
			queryTarget = target
		}
		input, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.source.AssignmentID, LeaseToken: poll.Assignment.LeaseToken, View: view, Target: queryTarget, RequestID: "parent-cli"})
		r := httptest.NewRequest(http.MethodPost, "/v1/agents/"+f.source.AgentID+"/assignment-context", strings.NewReader(string(input)))
		r.Header.Set(deviceIDHeader, "device-shared-studio")
		r.Header.Set(deviceSecretHeader, assignmentContextTestSecret)
		w := httptest.NewRecorder()
		f.s.Handler().ServeHTTP(w, r)
		if w.Code != 200 || !strings.Contains(w.Body.String(), target) || !strings.Contains(w.Body.String(), `"live":true`) {
			t.Fatalf("parent CLI %s: %d %s", view, w.Code, w.Body.String())
		}
		if view == AssignmentContextViewAgent {
			var out struct{ Data AgentDirectoryDetail }
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || !out.Data.Agent.Callable || out.Data.Agent.Description == "" || out.Data.Agent.RoleStatus != AgentDirectoryRoleDescribed {
				t.Fatalf("missing current selection evidence: %s", w.Body.String())
			}
		}
	}
}
