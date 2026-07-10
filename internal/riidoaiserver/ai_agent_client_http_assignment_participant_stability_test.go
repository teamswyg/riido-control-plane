package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAssignmentKeepsParticipantWhileDurablyQueued(t *testing.T) {
	const token = "owner-token"
	const taskID = "task-participant-stability"
	const agentID = "agent-public-openclaw"
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1", Token: token, Scopes: []string{"ai-agent:*", "agent:*:poll"},
	}})

	bytes := aiAgentSmokeRequest(t, server, http.MethodPost,
		base+"/tasks/"+taskID+"/agent-assignments", token, `{"agent_id":"`+agentID+`"}`, http.StatusAccepted)
	var assigned AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, bytes, &assigned)
	if assigned.AssignmentID == "" || assigned.ThreadID == "" || assigned.RunID == "" ||
		assigned.AgentID != agentID || assigned.AssignmentState != AgentAssignmentStateRunning ||
		assigned.WorkStatus != AgentWorkStatusRunning || assigned.ActiveStream == nil {
		t.Fatalf("accepted assignment lost participant identity: %+v", assigned)
	}

	bytes = aiAgentSmokeRequest(t, server, http.MethodGet,
		base+"/tasks/"+taskID+"/assignable-agents", token, "", http.StatusOK)
	var assignable AgentClientListResponse
	aiAgentSmokeDecode(t, bytes, &assignable)
	if !participantAgentListContains(assignable.Agents, agentID) {
		t.Fatalf("assignable agents removed active agent: %+v", assignable.Agents)
	}

	bytes = aiAgentSmokeRequest(t, server, http.MethodGet,
		"/v3/client/workspaces/"+defaultAIAgentClientWorkspaceID+"/ai-agent/tasks/"+taskID+"/threads",
		token, "", http.StatusOK)
	var history AIAgentTaskThreadHistoryCollectionResponse
	aiAgentSmokeDecode(t, bytes, &history)
	thread := historyThreadByID(t, history, assigned.ThreadID)
	if thread.AssignmentID != assigned.AssignmentID || thread.AssignmentState != AgentAssignmentStateQueued ||
		thread.ActiveStream == nil || history.ActiveStream == nil {
		t.Fatalf("v3 history lost active queued participant: %+v", thread)
	}
}

func participantAgentListContains(agents []AgentClientRecord, agentID string) bool {
	for _, agent := range agents {
		if agent.AgentID == agentID {
			return true
		}
	}
	return false
}
