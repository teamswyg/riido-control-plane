package riidoaiserver

import "maps"

func (s *Store) commitCLIInstallation(state *storeState, msg cliInstallationCommand) cliInstallationResult {
	if err := msg.ctx.Err(); err != nil {
		return cliInstallationResult{err: err}
	}
	if msg.read != nil {
		return s.readLocalCLIInstallation(state, *msg.read)
	}
	if msg.receipt != nil {
		return s.readLocalCLIInstallationReceipt(state, *msg.receipt)
	}
	if msg.replay != nil {
		return s.commitCLIInstallationIntentReplay(state, msg.ctx, *msg.replay)
	}
	trial := *state
	trial.cliInstallationIntents = maps.Clone(state.cliInstallationIntents)
	trial.cliInstallationLinks = maps.Clone(state.cliInstallationLinks)
	trial.devicePolicyRevisions = maps.Clone(state.devicePolicyRevisions)
	var result cliInstallationResult
	switch {
	case msg.create != nil:
		intent := *msg.create
		if err := intent.valid(); err != nil {
			return cliInstallationResult{err: err}
		}
		if old, ok := trial.cliInstallationIntents[intent.ID]; ok && old != intent {
			return cliInstallationResult{err: errCLIInstallationConflict}
		}
		trial.cliInstallationIntents[intent.ID] = intent
	case msg.confirm != nil:
		result.link, result.err = s.confirmLocalCLIInstallation(&trial, *msg.confirm)
		if result.err != nil {
			return result
		}
	default:
		return cliInstallationResult{err: errCLIInstallationInvalid}
	}
	if s.snapshotStore != nil {
		if err := s.snapshotStore.SaveStoreSnapshot(msg.ctx, snapshotFromState(&trial, s.now())); err != nil {
			return cliInstallationResult{err: err}
		}
	}
	state.cliInstallationIntents, state.cliInstallationLinks, state.devicePolicyRevisions = trial.cliInstallationIntents, trial.cliInstallationLinks, trial.devicePolicyRevisions
	return result
}
