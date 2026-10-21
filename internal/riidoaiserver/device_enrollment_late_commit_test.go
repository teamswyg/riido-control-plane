package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

type enrollmentLateCommitSnapshotStore struct {
	AIAgentClientSnapshotStore
	delay          bool
	pending        *AIAgentClientSnapshot
	pendingContext context.Context
	committed      AIAgentClientSnapshot
}

func (s *enrollmentLateCommitSnapshotStore) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if s.delay {
		s.delay, s.pending, s.pendingContext = false, &snapshot, ctx
		return errors.New("injected write still in flight")
	}
	if s.pending != nil {
		if err := s.AIAgentClientSnapshotStore.SaveAIAgentClientSnapshot(s.pendingContext, *s.pending); err != nil {
			return err
		}
		s.committed, s.pending = *s.pending, nil
	}
	return s.AIAgentClientSnapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot)
}

func TestDynamoEnrollmentLateCommitSurvivesOlderReadbackAndOrdinarySave(t *testing.T) {
	ctx := context.Background()
	fixture := newMembershipSharedDynamoFixture(t)
	native := newSnapshotDynamoDBStore(t, time.Now(), fixture.server.URL, nil)
	defer native.Close()
	outcome := &enrollmentLateCommitSnapshotStore{AIAgentClientSnapshotStore: native}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), outcome)
	if err != nil {
		t.Fatal(err)
	}
	principal := profileColorTestPrincipal()
	req := EnrollDeviceRequest{MachineID: "late-native-commit"}
	old, err := store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
	if err != nil {
		t.Fatal(err)
	}
	outcome.delay = true
	if _, err := store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req); err == nil {
		t.Fatal("older readback incorrectly proved commit")
	}
	// The forced reload still sees old state. The prior enrollment lands after
	// that read but before the ordinary writer's atomic CURRENT/part read.
	if _, err := store.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "must preserve late credential"}); err != nil {
		t.Fatal(err)
	}
	latest, found, err := native.LoadAIAgentClientSnapshot(ctx)
	if err != nil || !found || !sameEnrollmentCredential(latest, outcome.committed, old.DeviceID) {
		t.Fatal("ordinary writer overwrote a late committed enrollment")
	}
	fresh, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), native)
	if err != nil {
		t.Fatal(err)
	}
	assertEnrollmentCredential(t, fresh, old, false)
}
