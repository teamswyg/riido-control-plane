package riidoaiserver

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/hostintegration"
)

type cancellableRecoveryInventory struct {
	*DevelopmentAIAgentClientStore
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (c *cancellableRecoveryInventory) ListAIAgentTaskAssignableAgents(ctx context.Context, p AuthorizationResult, task string) (AgentClientListResponse, error) {
	c.once.Do(func() { close(c.entered) })
	select {
	case <-ctx.Done():
		return AgentClientListResponse{}, ctx.Err()
	case <-c.release:
	}
	return c.DevelopmentAIAgentClientStore.ListAIAgentTaskAssignableAgents(ctx, p, task)
}

func preflightFixture(t *testing.T) (*Store, *conditionalTestOperations, *cancellableRecoveryInventory, AgentEventRequest) {
	t.Helper()
	base := NewDevelopmentAIAgentClientStore()
	binding, _ := base.LookupDaemonBinding("agent-owned-codex")
	base.devices[0].DaemonLastSeenAt = time.Now()
	for i := range base.devices[0].Runtimes {
		r := &base.devices[0].Runtimes[i]
		r.DaemonID = binding.DaemonID
		r.LastDetectedAt = time.Now()
		r.HealthStatus = hostintegration.ProviderHealthHealthy
		r.DiagnosticCode = hostintegration.ProviderDiagnosticNone
	}
	client := &cancellableRecoveryInventory{DevelopmentAIAgentClientStore: base, entered: make(chan struct{}), release: make(chan struct{})}
	a := Assignment{ID: "preflight", TaskID: "task", AgentID: binding.AgentID, CreatedBy: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID, RuntimeProvider: "codex", State: AssignmentRunning, LeaseToken: "lease", UpdatedAt: time.Now(), RecoveryPolicy: recoveryTestPolicy()}
	ops := &conditionalTestOperations{projection: AssignmentProjection{Assignment: a, LastEventSeq: 10}}
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops, AgentRegistry: client})
	t.Cleanup(func() { store.Close() })
	req := recoveryEventRequest(a)
	req.DaemonID, req.DeviceID, req.RuntimeID = binding.DaemonID, binding.DeviceID, binding.RuntimeID
	req.Metadata[assignment.RecoveryProviderKey] = "cursor"
	return store, ops, client, req
}
