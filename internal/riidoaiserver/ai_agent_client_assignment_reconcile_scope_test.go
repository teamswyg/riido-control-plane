package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHTTPAIAgentClientAssignReconcilesRequestedTaskOnly(t *testing.T) {
	server, aiAgentStore := newRecordingAIAgentAssignmentTestServer(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/client/ai-agent/tasks/task-reconcile-scope/assignment",
		strings.NewReader(`{"agent_id":"agent-public-openclaw"}`),
	)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", resp.Code, resp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	if assigned.TaskID != "task-reconcile-scope" || assigned.AssignmentID == "" {
		t.Fatalf("assign response = %+v", assigned)
	}
	assertReconcileTaskIDs(t, aiAgentStore.reconcileTaskIDs(), "task-reconcile-scope")
}

func TestHTTPAIAgentClientCreateTaskAgentAssignmentReconcilesRequestedTaskOnly(t *testing.T) {
	server, aiAgentStore := newRecordingAIAgentAssignmentTestServer(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-additive-reconcile-scope/agent-assignments",
		strings.NewReader(`{"agent_id":"agent-public-openclaw"}`),
	)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("create agent assignment status=%d body=%s", resp.Code, resp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("create agent assignment json: %v", err)
	}
	if assigned.TaskID != "task-additive-reconcile-scope" || assigned.AssignmentID == "" {
		t.Fatalf("create agent assignment response = %+v", assigned)
	}
	assertReconcileTaskIDs(t, aiAgentStore.reconcileTaskIDs(), "task-additive-reconcile-scope")
}

func TestAIAgentClientProjectionReconcileTaskScopeSkipsUnrelatedActiveThreads(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	currentAgent, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "current scope agent",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
		ModelID:    stringPtr("cursor-auto"),
	})
	if err != nil {
		t.Fatalf("CreateAIAgent current: %v", err)
	}
	current, err := store.AssignAIAgentTask(ctx, principal, "task-current-scope", AssignAIAgentTaskRequest{
		AgentID:      currentAgent.Agent.AgentID,
		AssignmentID: "asn-current-scope",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask current: %v", err)
	}
	other, err := store.AssignAIAgentTask(ctx, principal, "task-other-scope", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-other-scope",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask other: %v", err)
	}
	now := time.Date(2026, 6, 9, 7, 0, 0, 0, time.UTC)
	reader := &recordingAssignmentProjectionReader{projections: map[string]AssignmentProjection{
		current.AssignmentID: {
			Assignment: Assignment{
				ID:              current.AssignmentID,
				TaskID:          current.TaskID,
				AgentID:         current.AgentID,
				RuntimeProvider: "openclaw",
				State:           AssignmentQueued,
				UpdatedAt:       now,
			},
		},
		other.AssignmentID: {
			Assignment: Assignment{
				ID:              other.AssignmentID,
				TaskID:          other.TaskID,
				AgentID:         other.AgentID,
				RuntimeProvider: "codex",
				State:           AssignmentFailed,
				UpdatedAt:       now,
			},
		},
	}}

	changed, err := store.ReconcileAIAgentActiveThreadProjections(ctx, principal, current.TaskID, reader)
	if err != nil {
		t.Fatalf("ReconcileAIAgentActiveThreadProjections: %v", err)
	}
	if !changed {
		t.Fatal("expected current task projection to update")
	}
	assertLoadedAssignmentIDs(t, reader.loadedAssignmentIDs(), current.AssignmentID)

	currentThreads, err := store.ListAIAgentTaskThreads(ctx, principal, current.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads current: %v", err)
	}
	if currentThreads.ActiveStream == nil {
		t.Fatalf("current task should keep active stream for queued thread: %+v", currentThreads)
	}
	if currentThreads.Threads[0].AssignmentState != AgentAssignmentStateQueued || currentThreads.Threads[0].WorkStatus != AgentWorkStatusIdle {
		t.Fatalf("client projection should preserve queued lifecycle without working copy: %+v", currentThreads)
	}
	store.mu.Lock()
	durableCurrent := store.taskThreads[current.TaskID][0]
	store.mu.Unlock()
	if durableCurrent.AssignmentState != AgentAssignmentStateQueued {
		t.Fatalf("durable thread should remain queued: %+v", durableCurrent)
	}
	otherThreads, err := store.ListAIAgentTaskThreads(ctx, principal, other.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads other: %v", err)
	}
	if otherThreads.ActiveStream == nil || otherThreads.Threads[0].AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("unrelated task should stay untouched: %+v", otherThreads)
	}
}

func TestAIAgentClientTerminalProjectionDoesNotExposeActiveParticipant(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 9, 7, 30, 0, 0, time.UTC)
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-terminal-projection", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-terminal-projection",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	reader := mapAssignmentProjectionReader{projections: map[string]AssignmentProjection{
		assigned.AssignmentID: {
			Assignment: Assignment{
				ID:              assigned.AssignmentID,
				TaskID:          assigned.TaskID,
				AgentID:         assigned.AgentID,
				RuntimeProvider: "openclaw",
				State:           AssignmentFailed,
				UpdatedAt:       now,
			},
		},
	}}
	changed, err := store.ReconcileAIAgentActiveThreadProjections(ctx, principal, assigned.TaskID, reader)
	if err != nil {
		t.Fatalf("ReconcileAIAgentActiveThreadProjections: %v", err)
	}
	if !changed {
		t.Fatal("expected failed projection to update stale active thread")
	}

	threads, err := store.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if threads.ActiveStream != nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateFailed ||
		threads.Threads[0].CompletedAt.IsZero() {
		t.Fatalf("terminal projection should close active participant thread: %+v", threads)
	}

	profiles, err := store.ListWorkspaceAssignedAgentProfiles(ctx, principal)
	if err != nil {
		t.Fatalf("ListWorkspaceAssignedAgentProfiles: %v", err)
	}
	if _, ok := profiles.AssignedAgentProfiles[assigned.TaskID]; ok {
		t.Fatalf("terminal projection leaked into assigned agent profiles: %+v", profiles.AssignedAgentProfiles)
	}
}

type recordingAIAgentClientStore struct {
	*DevelopmentAIAgentClientStore
	recordingMu        sync.Mutex
	reconcileTaskIDLog []string
}

func (s *recordingAIAgentClientStore) ReconcileAIAgentActiveThreadProjections(
	ctx context.Context,
	principal AuthorizationResult,
	taskID string,
	reader AssignmentProjectionReader,
) (bool, error) {
	s.recordingMu.Lock()
	s.reconcileTaskIDLog = append(s.reconcileTaskIDLog, strings.TrimSpace(taskID))
	s.recordingMu.Unlock()
	return s.DevelopmentAIAgentClientStore.ReconcileAIAgentActiveThreadProjections(ctx, principal, taskID, reader)
}

func (s *recordingAIAgentClientStore) reconcileTaskIDs() []string {
	s.recordingMu.Lock()
	defer s.recordingMu.Unlock()
	return append([]string(nil), s.reconcileTaskIDLog...)
}

type recordingAssignmentProjectionReader struct {
	projections map[string]AssignmentProjection
	mu          sync.Mutex
	loadedIDs   []string
}

func (r *recordingAssignmentProjectionReader) LoadAssignmentProjection(_ context.Context, assignmentID string) (AssignmentProjection, bool, error) {
	r.mu.Lock()
	r.loadedIDs = append(r.loadedIDs, strings.TrimSpace(assignmentID))
	r.mu.Unlock()
	projection, ok := r.projections[assignmentID]
	return projection, ok, nil
}

func (r *recordingAssignmentProjectionReader) loadedAssignmentIDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.loadedIDs...)
}

func newRecordingAIAgentAssignmentTestServer(t *testing.T) (http.Handler, *recordingAIAgentClientStore) {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := &recordingAIAgentClientStore{DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore()}
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	return NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler(), aiAgentStore
}

func assertReconcileTaskIDs(t *testing.T, got []string, want ...string) {
	t.Helper()
	assertStringSequence(t, "reconcile task ids", got, want...)
	for _, taskID := range got {
		if strings.TrimSpace(taskID) == "" {
			t.Fatalf("reconcile should not run workspace-wide from assign path: %v", got)
		}
	}
}

func assertLoadedAssignmentIDs(t *testing.T, got []string, want ...string) {
	t.Helper()
	assertStringSequence(t, "loaded assignment ids", got, want...)
}

func assertStringSequence(t *testing.T, label string, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", label, got, want)
		}
	}
}
