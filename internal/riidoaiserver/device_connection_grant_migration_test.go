package riidoaiserver

import (
	"context"
	"testing"
)

func TestLegacySnapshotBackfillsOwnerConnectionGrant(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	writer, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	enrollment, err := writer.EnrollDeviceCredential(ctx,
		AuthorizationResult{PrincipalID: "legacy-owner", WorkspaceID: "legacy-workspace"},
		"legacy-workspace",
		EnrollDeviceRequest{MachineID: "legacy-machine"},
	)
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	snapshots.snapshot.DeviceConnectionGrants = nil

	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("reopen legacy snapshot: %v", err)
	}
	reopened.mu.Lock()
	revision, principalCount := reopened.deviceConnectionSummaryLocked(enrollment.DeviceID)
	reopened.mu.Unlock()
	if principalCount != 1 || revision == "" {
		t.Fatalf("backfilled connection summary count=%d revision=%q", principalCount, revision)
	}
}
