package riidoaiserver

import (
	"context"
	"time"
)

func OpenStoreWithConfig(ctx context.Context, config StoreConfig) (*Store, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	now := config.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	activeLeaseDuration := config.ActiveLeaseDuration
	if activeLeaseDuration <= 0 {
		activeLeaseDuration = time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second
	}
	state := newStoreState()
	loadedSnapshot := false
	if config.SnapshotStore != nil {
		snapshot, ok, err := config.SnapshotStore.LoadStoreSnapshot(ctx)
		if err != nil {
			return nil, err
		}
		if ok {
			loaded, err := stateFromSnapshot(snapshot)
			if err != nil {
				return nil, err
			}
			state = loaded
			loadedSnapshot = true
		}
	}
	if !loadedSnapshot {
		if loader, ok := config.OperationStore.(AssignmentOperationLoader); ok {
			operations, err := loader.LoadAssignmentOperations(ctx)
			if err != nil {
				return nil, err
			}
			if len(operations) > 0 {
				loaded, err := stateFromAssignmentOperations(operations)
				if err != nil {
					return nil, err
				}
				state = loaded
			}
		}
	}
	if reader, ok := config.OperationStore.(AssignmentProjectionReader); ok {
		projections, err := loadReplayAssignmentProjections(ctx, &state, reader)
		if err != nil {
			return nil, err
		}
		overlayAssignmentProjections(&state, projections)
	}
	repairs := repairStaleReplayAssignments(&state, now().UTC(), activeLeaseDuration)
	if config.OperationStore != nil {
		for _, repair := range repairs {
			if err := config.OperationStore.SaveAssignmentOperation(ctx, repair); err != nil {
				return nil, err
			}
		}
	}
	return newStoreWithConfig(config, state), nil
}
