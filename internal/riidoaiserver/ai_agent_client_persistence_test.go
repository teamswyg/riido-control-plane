package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
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
