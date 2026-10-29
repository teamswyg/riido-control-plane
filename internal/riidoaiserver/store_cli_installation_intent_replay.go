package riidoaiserver

import (
	"context"
	"maps"
)

func (s *Store) createOrReplayCLIInstallationIntent(ctx context.Context, candidate cliInstallationIntentReplay) (cliInstallationIntentReplayResponse, error) {
	if s.snapshotStore != nil && s.operationStore != nil {
		return cliInstallationIntentReplayResponse{}, errCLIInstallationUnavailable
	}
	if durable, ok := s.operationStore.(cliInstallationRepository); ok {
		return durable.createOrReplayCLIInstallationIntent(ctx, candidate)
	}
	result, err := s.executeCLIInstallation(ctx, cliInstallationCommand{replay: &candidate})
	return result.replay, err
}

func (s *Store) commitCLIInstallationIntentReplay(state *storeState, ctx context.Context, candidate cliInstallationIntentReplay) cliInstallationResult {
	if err := candidate.valid(); err != nil {
		return cliInstallationResult{err: err}
	}
	key := candidate.key()
	if original, ok := state.cliInstallationReplays[key]; ok {
		intent, found := state.cliInstallationIntents[original.Response.IntentID]
		if original.valid() != nil || original.key() != key || !found || intent.valid() != nil || !original.matches(intent) {
			return cliInstallationResult{err: errCLIInstallationUnavailable}
		}
		if original.Request != candidate.Request || original.Response.RequestID != candidate.Response.RequestID {
			return cliInstallationResult{err: errCLIInstallationConflict}
		}
		// Save may have committed before returning an error. The actor keeps
		// that original pair even on failure; no retry may ACK it until the
		// complete current snapshot has been persisted successfully.
		if s.snapshotStore != nil {
			if err := s.snapshotStore.SaveStoreSnapshot(ctx, snapshotFromState(state, s.now())); err != nil {
				return cliInstallationResult{err: err}
			}
		}
		return cliInstallationResult{replay: original.Response}
	}
	intent := candidate.intent()
	if _, exists := state.cliInstallationIntents[intent.ID]; exists {
		return cliInstallationResult{err: errCLIInstallationUnavailable}
	}
	trial := *state
	trial.cliInstallationIntents, trial.cliInstallationReplays = maps.Clone(state.cliInstallationIntents), maps.Clone(state.cliInstallationReplays)
	trial.cliInstallationIntents[intent.ID], trial.cliInstallationReplays[key] = intent, candidate
	// Retain the immutable pair before attempting persistence. A store error
	// does not prove that the atomic file write was rolled back. Keeping this
	// pair in the complete actor state prevents any later snapshot-producing
	// command from overwriting the first nonce, even if readback is unavailable.
	// The failing request receives no nonce; a replay must save successfully.
	state.cliInstallationIntents, state.cliInstallationReplays = trial.cliInstallationIntents, trial.cliInstallationReplays
	if s.snapshotStore != nil {
		if err := s.snapshotStore.SaveStoreSnapshot(ctx, snapshotFromState(state, s.now())); err != nil {
			return cliInstallationResult{err: err}
		}
	}
	return cliInstallationResult{replay: candidate.Response}
}
