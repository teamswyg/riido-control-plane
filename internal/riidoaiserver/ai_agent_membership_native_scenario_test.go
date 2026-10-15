package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type membershipNativeScenario struct {
	f                            *membershipSharedDynamoFixture
	client                       *PersistentAIAgentClientStore
	store                        *Store
	handler                      http.Handler
	config                       ServerConfig
	taskID, agentID, token, base string
}

func newMembershipNativeScenario(t *testing.T, taskID string) *membershipNativeScenario {
	t.Helper()
	f := newMembershipSharedDynamoFixture(t)
	client, store := newMembershipDynamoReplica(t, f)
	s := &membershipNativeScenario{f: f, client: client, store: store, taskID: taskID, agentID: "agent-public-openclaw", token: "ai-agent-token", base: "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"}
	s.config = ServerConfig{AIAgentClient: client, Assignment: store, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}
	s.handler = NewServer(s.config).Handler()
	return s
}

func (s *membershipNativeScenario) replicaHandler(client AIAgentClientStore, store AssignmentStore) http.Handler {
	config := s.config
	config.AIAgentClient, config.Assignment = client, store
	return NewServer(config).Handler()
}

func (s *membershipNativeScenario) post(handler http.Handler, suffix, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, s.base+"/tasks/"+s.taskID+suffix, strings.NewReader(body))
	request.Header.Set(aiAgentTokenHeader, s.token)
	result := httptest.NewRecorder()
	handler.ServeHTTP(result, request)
	return result
}

func (s *membershipNativeScenario) stoppedAssignment(t *testing.T) AIAgentTaskActionResponse {
	t.Helper()
	first := createProviderMultiAssignment(t, s.handler, s.base, s.token, s.taskID, s.agentID)
	stopped := s.post(s.handler, "/agent-assignments/"+s.agentID+"/stop", `{"assignment_id":"`+first.AssignmentID+`"}`)
	if stopped.Code != http.StatusAccepted {
		t.Fatalf("initial stop=%d %s", stopped.Code, stopped.Body.String())
	}
	return first
}

func (s *membershipNativeScenario) participants(t *testing.T) []TaskAgentParticipantRecord {
	t.Helper()
	profiles, err := s.client.ListWorkspaceAssignedAgentProfiles(context.Background(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID})
	if err != nil {
		t.Fatal(err)
	}
	return profiles.TaskAgentParticipants[s.taskID].Agents
}
