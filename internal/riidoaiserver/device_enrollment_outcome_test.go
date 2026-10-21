package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

type enrollmentOutcomeSnapshotStore struct {
	AIAgentClientSnapshotStore
	loseResponse, failRead bool
}

func (s *enrollmentOutcomeSnapshotStore) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if err := s.AIAgentClientSnapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot); err != nil {
		return err
	}
	if s.loseResponse {
		return errors.New("injected lost commit response")
	}
	return nil
}

func (s *enrollmentOutcomeSnapshotStore) LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error) {
	if s.failRead {
		return AIAgentClientSnapshot{}, false, errors.New("injected unavailable readback")
	}
	return s.AIAgentClientSnapshotStore.LoadAIAgentClientSnapshot(ctx)
}

func TestEnrollmentResolvesCommitResponseLossAndFencesUnknownReadback(t *testing.T) {
	for _, failedRead := range []bool{false, true} {
		t.Run(map[bool]string{false: "confirmed", true: "unknown"}[failedRead], func(t *testing.T) {
			ctx := context.Background()
			principal := profileColorTestPrincipal()
			memory := &memoryAIAgentClientSnapshotStore{}
			outcome := &enrollmentOutcomeSnapshotStore{AIAgentClientSnapshotStore: memory}
			store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), outcome)
			if err != nil {
				t.Fatal(err)
			}
			req := EnrollDeviceRequest{MachineID: "ambiguous-enrollment"}
			first, err := store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
			if err != nil {
				t.Fatal(err)
			}
			outcome.loseResponse, outcome.failRead = true, failedRead
			second, err := store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
			if (err != nil) != failedRead {
				t.Fatalf("readback outcome error=%v", err)
			}
			if !failedRead {
				assertEnrollmentCredential(t, store, second, true)
				assertEnrollmentCredential(t, store, first, false)
				return
			}
			if _, err := store.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "must not write stale state"}); err == nil {
				t.Fatal("unknown readback allowed a stale write")
			}
			if _, err := store.AuthorizeDeviceCredential(ctx, first.DeviceID, first.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent}); err == nil {
				t.Fatal("unknown readback authorized stale credential")
			}
			outcome.loseResponse, outcome.failRead = false, false
			if _, err := store.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "fresh read recovers"}); err != nil {
				t.Fatal(err)
			}
			assertEnrollmentCredential(t, store, first, false)
			if !sameEnrollmentCredential(memory.snapshot, mustEnrollmentSnapshot(t, store), first.DeviceID) {
				t.Fatal("fresh read overwrote the committed credential")
			}
		})
	}
}

func mustEnrollmentSnapshot(t *testing.T, s *PersistentAIAgentClientStore) AIAgentClientSnapshot {
	t.Helper()
	snapshot, err := s.snapshot(s.snapshotNow())
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
