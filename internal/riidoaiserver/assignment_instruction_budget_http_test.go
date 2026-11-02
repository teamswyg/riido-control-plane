package riidoaiserver

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestHTTPProfileInstructionBudgetThroughAssignmentPoll(t *testing.T) {
	for _, n := range []int{798, 1000, 1001} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
				PrincipalID: "user-1", Token: "owner", Scopes: []string{"ai-agent:*", "agent:*:poll"},
			}})
			// Include a non-BMP character to distinguish characters from bytes and UTF-16 units.
			raw := strings.Repeat("지", n-1) + "🦊"
			base := "/v2/client/workspaces/workspace-a/ai-agent"
			body := aiAgentSmokeJSON(t, CreateAgentConfigurationRequest{
				Name: "Instruction budget", RuntimeID: "runtime-cursor-dev", Visibility: AgentVisibilityPublic, Instruction: &raw,
			})
			if n > AgentInstructionMaxCharacters {
				aiAgentSmokeRequest(t, server, http.MethodPost, base+"/agents", "owner", body, http.StatusBadRequest)
				return
			}
			var created AgentClientRecordResponse
			aiAgentSmokeDecode(t, aiAgentSmokeRequest(t, server, http.MethodPost, base+"/agents", "owner", body, http.StatusCreated), &created)
			if created.Agent.Instruction != raw {
				t.Fatal("saved profile instruction changed")
			}
			aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/task-v2/agent-assignments", "owner",
				`{"agent_id":"`+created.Agent.AgentID+`","request_id":"instruction-boundary"}`, http.StatusAccepted)
			var poll PollResponse
			aiAgentSmokeDecode(t, aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/agents/"+created.Agent.AgentID+"/poll", "owner",
				`{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-cursor-dev"}`, http.StatusOK), &poll)
			if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.AgentInstruction != raw+"\n\n"+agentResponseGuidelines {
				t.Fatal("daemon did not receive the complete profile and response guidelines exactly once")
			}
		})
	}
}

func TestHTTPDirectAssignmentValidatesCallerInstruction(t *testing.T) {
	for _, n := range []int{1000, 1001} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
				PrincipalID: "user-1", Token: "owner", Scopes: []string{"riido:*"},
			}})
			body := aiAgentSmokeJSON(t, AssignRequest{
				ComponentID: "component-a", AgentID: "agent-owned-codex",
				RuntimeProvider: "codex", Prompt: "test budget", AgentInstruction: strings.Repeat("지", n),
			})
			want := http.StatusOK
			if n > AgentInstructionMaxCharacters {
				want = http.StatusBadRequest
			}
			response := aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/component-tasks/component-a/assignment", "owner", body, want)
			if want == http.StatusBadRequest && !strings.Contains(string(response), "agent_instruction must be 1000 characters or fewer") {
				t.Fatalf("unexpected rejection: %s", response)
			}
		})
	}
}
