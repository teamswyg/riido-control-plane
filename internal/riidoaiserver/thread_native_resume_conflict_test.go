package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mismatchedNativeRuntimeStore struct {
	AIAgentClientStore
	binding AgentRuntimeBinding
	runtime RuntimeRecord
}

func (s mismatchedNativeRuntimeStore) LookupAgentRuntimeFact(string) (AgentRuntimeBinding, RuntimeRecord, bool) {
	return s.binding, s.runtime, true
}

func TestNativeFollowupMismatchIsNotAnActiveThreadConflict(t *testing.T) {
	for _, name := range []string{"model", "provider", "worktree", "runtime"} {
		t.Run(name, func(t *testing.T) {
			s, _, thread, principal, req := nativeResumeScopeFixture(t)
			kind := "session_configuration_mismatch"
			message := "native session configuration does not match the requested agent, provider, model or worktree"
			switch name {
			case "model":
				req.ModelID = "different-model"
			case "provider":
				req.RuntimeProvider = "claude"
			case "worktree":
				req.Worktree = &AssignmentWorktree{BranchName: "different-branch"}
			case "runtime":
				binding, runtime, err := s.resolveAgentRuntimeFact(req.AgentID)
				if err != nil {
					t.Fatal(err)
				}
				runtime.RuntimeID = "different-runtime"
				s.aiAgent = mismatchedNativeRuntimeStore{s.aiAgent, binding, runtime}
				kind = "session_runtime_mismatch"
				message = "native session runtime binding is inconsistent"
			}
			_, err := s.applyThreadNativeResume(t.Context(), principal, thread, &req, "user-1")
			if err == nil || req.ResumeSessionID != "" {
				t.Fatal("mismatched session was accepted")
			}
			if errors.Is(err, ErrAIAgentTaskThreadConflict) {
				t.Fatal("terminal session mismatch was reported as an active thread conflict")
			}
			wrapped := withFailureStage("native_resume", err)
			if got := failureKind(t.Context(), wrapped); got != kind {
				t.Fatalf("failure kind = %q, want %q", got, kind)
			}
			response := httptest.NewRecorder()
			writeAIAgentClientError(response, wrapped)
			if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), message) {
				t.Fatal("session mismatch lost its conflict status or specific explanation")
			}
		})
	}
}

func TestHTTPCompletedThreadModelMismatchReportsSessionConflict(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	f.client.mu.Lock()
	agent := f.client.agents[first.AgentID]
	agent.ModelID = "different-model"
	f.client.agents[first.AgentID] = agent
	f.client.mu.Unlock()
	req := httptest.NewRequest(http.MethodPost, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", strings.NewReader(`{"body":"Continue the task"}`))
	req.Header.Set(aiAgentTokenHeader, f.token)
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, req)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "native session configuration does not match") || strings.Contains(response.Body.String(), ErrAIAgentTaskThreadConflict.Error()) {
		t.Fatal("completed thread model mismatch was not explained as a session configuration conflict")
	}
	projection, found, err := f.store.LoadAssignmentProjection(t.Context(), first.AssignmentID)
	if err != nil || !found || projection.Assignment.State != AssignmentCompleted {
		t.Fatal("rejected followup changed the prior execution")
	}
	poll, err := f.store.PollAgent(t.Context(), first.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
	if err != nil || poll.Assignment != nil {
		t.Fatal("rejected followup published an assignment")
	}
}
