package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func participantAssignOpenClaw(t *testing.T, server http.Handler, token string) AIAgentTaskActionResponse {
	t.Helper()
	body := `{"agent_id":"agent-public-openclaw"}`
	bytes := aiAgentSmokeRequest(t, server, http.MethodPost, "/v1/client/ai-agent/tasks/task-new/assignment", token, body, http.StatusAccepted)
	var assigned AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, bytes, &assigned)
	if assigned.TaskID != "task-new" ||
		assigned.AgentID != "agent-public-openclaw" ||
		assigned.AssignmentState != AgentAssignmentStateRunning ||
		assigned.CommentKind != AgentTaskCommentAssignmentStarted ||
		assigned.ThreadID == "" {
		t.Fatalf("assign response = %+v", assigned)
	}
	return assigned
}

func participantPostFollowup(t *testing.T, server http.Handler, token, threadID string) AIAgentTaskActionResponse {
	t.Helper()
	body := `{"body":"다음 작업을 이어서 진행해 주세요.","source_message_id":"message-next-1"}`
	path := "/v1/client/ai-agent/tasks/task-new/threads/" + threadID + "/messages"
	bytes := aiAgentSmokeRequest(t, server, http.MethodPost, path, token, body, http.StatusAccepted)
	var message AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, bytes, &message)
	return message
}

func participantUnassignOpenClaw(t *testing.T, server http.Handler, token string) AIAgentTaskActionResponse {
	t.Helper()
	body := `{"agent_id":"agent-public-openclaw","reason":"removed from participants"}`
	bytes := aiAgentSmokeRequest(t, server, http.MethodDelete, "/v1/client/ai-agent/tasks/task-new/assignment", token, body, http.StatusAccepted)
	var unassigned AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, bytes, &unassigned)
	return unassigned
}

func assertParticipantReplayOmitsStaleQueued(t *testing.T, server http.Handler, token string) {
	t.Helper()
	body := string(aiAgentSmokeRequest(t, server, http.MethodGet, "/v1/client/ai-agent/events?replay=1", token, "", http.StatusOK))
	if strings.Contains(body, string(AgentTaskCommentQueuedByBusyAgent)) ||
		!strings.Contains(body, string(AgentTaskCommentAssignmentStarted)) ||
		!strings.Contains(body, string(AgentTaskCommentStoppedByUserRequest)) {
		t.Fatalf("events body = %q", body)
	}
}
