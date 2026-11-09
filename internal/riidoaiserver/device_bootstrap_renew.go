package riidoaiserver

import (
	"context"
	"math"
)

func (s *PersistentAIAgentClientStore) renewDeviceBootstrap(ctx context.Context, id string, device deviceBootstrapAuth) (deviceBootstrapRecord, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	record, err := s.loadDeviceBootstrap(ctx, id, device)
	if err != nil {
		return deviceBootstrapRecord{}, err
	}
	if record.State == "complete" {
		return record, nil
	}
	// A live challenge is replayed exactly. Renewal never extends an already
	// issued challenge or clears a prepared admission fence.
	if s.snapshotNow().Before(record.Intent.ExpiresAt) {
		return record, nil
	}
	if record.Revision == math.MaxUint64 {
		return deviceBootstrapRecord{}, errCLIInstallationConflict
	}
	next := record
	next.Intent, _, err = newCLIInstallationIntent(record.Request.cliInstallationIntentRequest)
	if err != nil {
		return deviceBootstrapRecord{}, err
	}
	next.Revision++
	return s.commitDeviceBootstrap(ctx, deviceBootstrapMutation{before: &record, after: next})
}
