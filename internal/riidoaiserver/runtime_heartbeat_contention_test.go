package riidoaiserver

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSnapshotOperationCancelledWaiterDoesNotConsumeLock(t *testing.T) {
	store := &PersistentAIAgentClientStore{}
	store.operationMu.Lock()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := store.lockOperationContext(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancelled lock wait = %v", err)
	}
	store.operationMu.Unlock()
	if err := store.lockOperationContext(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expired caller consumed an available lock: %v", err)
	}
	next, nextCancel := context.WithTimeout(context.Background(), time.Second)
	defer nextCancel()
	if err := store.lockOperationContext(next); err != nil {
		t.Fatalf("cancelled waiter leaked operation lock: %v", err)
	}
	store.operationMu.Unlock()
}

// Workspace profile polling forces a durable membership read. A heartbeat must
// join those readers in the operation queue instead of repeatedly losing TryLock.
func TestRuntimeHeartbeatNotStarvedByWorkspaceProfileReads(t *testing.T) {
	ctx := context.Background()
	base := NewDevelopmentAIAgentClientStore()
	owner := AuthorizationResult{PrincipalID: "owner", WorkspaceID: "workspace"}
	enrollment := enrollTestDevice(t, ctx, base, owner, owner.WorkspaceID, "machine", "Mac")
	snapshots := &slowMembershipSnapshotStore{delay: 5 * time.Millisecond}
	store, err := OpenPersistentAIAgentClientStore(ctx, base, snapshots)
	if err != nil {
		t.Fatal(err)
	}
	stop := make(chan struct{})
	var readers sync.WaitGroup
	for range 4 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				_, _ = store.ListWorkspaceAssignedAgentProfiles(ctx, owner)
			}
		}()
	}
	t.Cleanup(func() { close(stop); readers.Wait() })
	// Let ordinary mutex waiters enter starvation-mode handoff before the
	// heartbeat arrives. Each durable read is much shorter than its deadline.
	time.Sleep(30 * time.Millisecond)
	heartbeatCtx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()
	_, err = store.SyncAIAgentDaemonRuntimeSnapshot(heartbeatCtx, owner, DeviceRuntimeSnapshotSyncRequest{
		DeviceID: enrollment.DeviceID, DaemonID: "machine",
		Runtimes: []RuntimeSnapshotRecord{{RuntimeID: "machine:claude", Kind: RuntimeKindClaudeCode,
			Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected}},
	})
	if err != nil {
		t.Fatalf("healthy runtime heartbeat starved behind repeated profile reads: %v", err)
	}
}

type slowMembershipSnapshotStore struct {
	memoryAIAgentClientSnapshotStore
	delay time.Duration
}

func (*slowMembershipSnapshotStore) requiresFreshTaskAgentMembership() bool { return true }

func (s *slowMembershipSnapshotStore) LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error) {
	select {
	case <-ctx.Done():
		return AIAgentClientSnapshot{}, false, ctx.Err()
	case <-time.After(s.delay):
		return s.memoryAIAgentClientSnapshotStore.LoadAIAgentClientSnapshot(ctx)
	}
}
