package riidoaiserver

import "context"

func (s *Store) retireCLIInstallationIntent(ctx context.Context, req cliInstallationRetirementRequest) (cliInstallationRetirement, error) {
	if s.snapshotStore != nil && s.operationStore != nil {
		return cliInstallationRetirement{}, errCLIInstallationUnavailable
	}
	if durable, ok := s.operationStore.(cliInstallationRepository); ok {
		return durable.retireCLIInstallationIntent(ctx, req)
	}
	result, err := s.executeCLIInstallation(ctx, cliInstallationCommand{retire: &req})
	return result.retired, err
}

func (s *Store) commitCLIInstallationRetirement(state *storeState, ctx context.Context, req cliInstallationRetirementRequest) cliInstallationResult {
	intent, ok := state.cliInstallationIntents[req.IntentID]
	if !ok {
		return cliInstallationResult{err: ErrAuthorizationForbidden}
	}
	if err := checkCLIInstallationRetirement(req, intent, s.now()); err != nil {
		return cliInstallationResult{err: err}
	}
	result, err := cliInstallationRetirementResult(req, intent, state.cliInstallationLinks[req.key()])
	if err != nil {
		return cliInstallationResult{err: err}
	}
	if !intent.Consumed {
		intent.Retired = true
		if err := intent.valid(); err != nil {
			return cliInstallationResult{err: err}
		}
		// Save can commit before reporting an error. Retain the fence in actor
		// memory so another snapshot write cannot erase that terminal state.
		// Every retirement retry must persist it before acknowledging success.
		state.cliInstallationIntents[intent.ID] = intent
	}
	if s.snapshotStore != nil {
		if err := s.snapshotStore.SaveStoreSnapshot(ctx, snapshotFromState(state, s.now())); err != nil {
			return cliInstallationResult{err: err}
		}
	}
	return cliInstallationResult{retired: result}
}
