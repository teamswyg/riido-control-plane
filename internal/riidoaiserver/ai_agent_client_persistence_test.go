package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPersistentAIAgentClientStoreRestoresDevelopmentState(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	created, err := store.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:        "Persistent Agent",
		Visibility:  AgentVisibilityPrivate,
		RuntimeID:   "runtime-cursor-dev",
		Description: stringPtr("development persistence"),
		Instruction: stringPtr("persist this configuration"),
	})
	if err != nil {
		t.Fatalf("CreateAIAgent: %v", err)
	}
	enrollment, err := store.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{DisplayName: "Development Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential: %v", err)
	}
	if _, err := store.SubmitAIAgentTaskComment(ctx, principal, "task-persist", SubmitAIAgentTaskCommentRequest{
		AgentID: created.Agent.AgentID,
		Body:    "continue from persisted state",
	}); err != nil {
		t.Fatalf("SubmitAIAgentTaskComment: %v", err)
	}
	rawSnapshot, err := json.Marshal(snapshots.snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	if strings.Contains(string(rawSnapshot), enrollment.DeviceSecret) {
		t.Fatal("snapshot must not contain the raw one-time device secret")
	}

	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("reopen persistent store: %v", err)
	}
	bootstrap, err := reopened.BootstrapAIAgentClient(ctx, principal, ClientKindWeb)
	if err != nil {
		t.Fatalf("BootstrapAIAgentClient: %v", err)
	}
	if !agentListContains(bootstrap.Agents, created.Agent.AgentID) {
		t.Fatalf("reopened bootstrap missing created agent: %+v", bootstrap.Agents)
	}
	devicePrincipal, err := reopened.AuthorizeDeviceCredential(ctx, enrollment.DeviceID, enrollment.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent})
	if err != nil {
		t.Fatalf("AuthorizeDeviceCredential: %v", err)
	}
	if devicePrincipal.PrincipalID != principal.PrincipalID {
		t.Fatalf("device principal = %+v", devicePrincipal)
	}
	events, err := reopened.AIAgentClientEvents(ctx, principal)
	if err != nil {
		t.Fatalf("AIAgentClientEvents: %v", err)
	}
	var foundTypedWorkStatus bool
	for _, event := range events {
		if _, ok := event.Payload.(AgentWorkStatusChangedEvent); ok {
			foundTypedWorkStatus = true
		}
	}
	if !foundTypedWorkStatus {
		t.Fatalf("reopened events should restore typed work-status payloads: %+v", events)
	}
}

func TestAIAgentClientSnapshotRetainsRecentReplayEvents(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.mu.Lock()
	store.events = nil
	for i := 1; i <= aiAgentClientReplayEventLimit+25; i++ {
		store.events = append(store.events, ClientStreamEvent{
			Seq:       int64(i),
			EventType: AgentClientEventWorkStatusChanged,
			Payload: AgentWorkStatusChangedEvent{
				EventType:       AgentClientEventWorkStatusChanged,
				SchemaVersion:   SchemaVersion,
				AgentID:         "agent-owned-codex",
				TaskID:          "task-replay",
				ThreadID:        "thread-replay",
				RunID:           "run-replay",
				WorkStatus:      AgentWorkStatusRunning,
				AssignmentState: AgentAssignmentStateRunning,
				CommentKind:     AgentTaskCommentRuntimeProgress,
			},
		})
	}
	store.mu.Unlock()

	snapshot, err := store.snapshot(time.Date(2026, 6, 3, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.Events) != aiAgentClientReplayEventLimit {
		t.Fatalf("snapshot retained %d events, want %d", len(snapshot.Events), aiAgentClientReplayEventLimit)
	}
	if snapshot.Events[0].Seq != 26 || snapshot.Events[len(snapshot.Events)-1].Seq != int64(aiAgentClientReplayEventLimit+25) {
		t.Fatalf("snapshot retained seq range %d..%d", snapshot.Events[0].Seq, snapshot.Events[len(snapshot.Events)-1].Seq)
	}

	reopened := NewDevelopmentAIAgentClientStore()
	if err := reopened.restoreSnapshot(snapshot); err != nil {
		t.Fatalf("restoreSnapshot: %v", err)
	}
	reopened.mu.Lock()
	event := reopened.appendClientEventLocked(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{
		EventType:       AgentClientEventWorkStatusChanged,
		SchemaVersion:   SchemaVersion,
		AgentID:         "agent-owned-codex",
		TaskID:          "task-replay",
		ThreadID:        "thread-replay",
		RunID:           "run-replay",
		WorkStatus:      AgentWorkStatusCompleted,
		AssignmentState: AgentAssignmentStateCompleted,
		CommentKind:     AgentTaskCommentTaskCompleted,
	})
	reopened.mu.Unlock()
	if event.Seq != int64(aiAgentClientReplayEventLimit+26) {
		t.Fatalf("next replay event seq = %d, want %d", event.Seq, aiAgentClientReplayEventLimit+26)
	}
}

type memoryAIAgentClientSnapshotStore struct {
	snapshot AIAgentClientSnapshot
	ok       bool
	saves    int
}

func (s *memoryAIAgentClientSnapshotStore) LoadAIAgentClientSnapshot(context.Context) (AIAgentClientSnapshot, bool, error) {
	return s.snapshot, s.ok, nil
}

func (s *memoryAIAgentClientSnapshotStore) SaveAIAgentClientSnapshot(_ context.Context, snapshot AIAgentClientSnapshot) error {
	s.snapshot = snapshot
	s.ok = true
	s.saves++
	return nil
}

func agentListContains(agents []AgentClientRecord, agentID string) bool {
	for _, agent := range agents {
		if agent.AgentID == agentID {
			return true
		}
	}
	return false
}
