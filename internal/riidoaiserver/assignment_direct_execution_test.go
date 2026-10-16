package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAssignmentDirectExecutionPreservesRequest(t *testing.T) {
	for _, route := range []string{"/v1/client/ai-agent/tasks/task-copy/assignment", "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-copy/agent-assignments"} {
		for _, instruction := range []string{"", "A안과 B안의 장단점 표로 비교 부탁. 비용과 접근성이 기준."} {
			t.Run(route+"/"+instruction, func(t *testing.T) {
				handler, store := newIntentGateHTTPTestServer(t)
				body, err := json.Marshal(map[string]string{"agent_id": "agent-owned-codex", "instruction": instruction, "request_id": "direct-request"})
				if err != nil {
					t.Fatal(err)
				}
				assign := func() AIAgentTaskActionResponse {
					req := httptest.NewRequest(http.MethodPost, route, strings.NewReader(string(body)))
					req.Header.Set(aiAgentTokenHeader, "user-token")
					resp := httptest.NewRecorder()
					handler.ServeHTTP(resp, req)
					if resp.Code != http.StatusAccepted {
						t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
					}
					return decodeAIAgentTaskActionResponse(t, resp.Body.Bytes())
				}
				assigned := assign()
				if assigned.WorkStatus == AgentWorkStatusWaitingForUser || isIntentGateAssignmentID(assigned.AssignmentID) {
					t.Fatalf("request blocked before model execution: status=%s message=%q", assigned.WorkStatus, assigned.Message)
				}
				if retry := assign(); retry.AssignmentID != assigned.AssignmentID {
					t.Fatalf("duplicate request created another run: %s vs %s", retry.AssignmentID, assigned.AssignmentID)
				}
				poll, err := store.PollAgent(t.Context(), assigned.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
				if err != nil || poll.Action != PollStart || poll.Assignment == nil {
					t.Fatalf("not delivered: action=%v err=%v", poll.Action, err)
				}
				for _, want := range []string{marketingTaskContextFixture().Component.Title, marketingTaskContextFixture().Document.Content, instruction} {
					if !strings.Contains(poll.Assignment.Prompt, want) {
						t.Fatalf("request content lost: %q", want)
					}
				}
				if strings.Contains(poll.Assignment.Prompt, "do_not_create_deliverables_until_user_replies") || strings.Contains(poll.Assignment.Prompt, clientMessageNeedUserInput) {
					t.Fatal("prompt still blocks execution")
				}
			})
		}
	}
}
