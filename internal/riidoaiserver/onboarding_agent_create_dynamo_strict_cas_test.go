package riidoaiserver

import (
	"context"
	"sync"
	"testing"
)

// Interpose at the external persistence boundary, after the creator has loaded
// and prepared its trial. The competing writer then changes an ordinary profile.
type onboardingAgentCreateInterposedSnapshot struct {
	inner      AIAgentClientSnapshotStore
	once       sync.Once
	beforeSave func()
}

func (s *onboardingAgentCreateInterposedSnapshot) LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error) {
	return s.inner.LoadAIAgentClientSnapshot(ctx)
}

func (s *onboardingAgentCreateInterposedSnapshot) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	s.once.Do(s.beforeSave)
	return s.inner.SaveAIAgentClientSnapshot(ctx, snapshot)
}

func TestOnboardingAgentCreateDynamoStrictCASPreservesConcurrentOrdinaryProfile(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	creator, _ := newMembershipDynamoReplica(t, fixture)
	competitor, _ := newMembershipDynamoReplica(t, fixture)
	principal := profileColorTestPrincipal()
	var competingErr error
	creator.snapshotStore = &onboardingAgentCreateInterposedSnapshot{inner: creator.snapshotStore, beforeSave: func() {
		_, competingErr = competitor.UpdateAIAgentConfiguration(t.Context(), principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "authoritative concurrent profile"})
	}}
	created, err := creator.CreateAIAgentFromOnboardingFixture(withOnboardingAgentCreateRequestID(t.Context(), "strict-whole-snapshot-create"), principal, "riido_pm", profileColorTestRequest("strict creation"))
	if competingErr != nil || err != nil {
		t.Fatalf("concurrent ordinary update=%v creation=%v", competingErr, err)
	}
	fresh, _ := newMembershipDynamoReplica(t, fixture)
	if fresh.agents["agent-owned-claude"].Name != "authoritative concurrent profile" {
		t.Fatal("fixture creation overwrote an unrelated profile committed after its trial load")
	}
	if _, exists := fresh.agents[created.Agent.AgentID]; !exists {
		t.Fatal("creation disappeared after resolving the strict snapshot conflict")
	}
}
