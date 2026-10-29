package riidoaiserver

import "context"

func (s *Store) readCLIInstallationReceipt(ctx context.Context, req cliInstallationReceiptRequest) (cliInstallationReceipt, error) {
	if s.snapshotStore != nil && s.operationStore != nil {
		return cliInstallationReceipt{}, errCLIInstallationUnavailable
	}
	if durable, ok := s.operationStore.(cliInstallationRepository); ok {
		return durable.readCLIInstallationReceipt(ctx, req)
	}
	result, err := s.executeCLIInstallation(ctx, cliInstallationCommand{receipt: &req})
	return result.receipt, err
}

func (s *Store) readLocalCLIInstallationReceipt(state *storeState, req cliInstallationReceiptRequest) cliInstallationResult {
	intent, ok := state.cliInstallationIntents[req.IntentID]
	if !ok {
		return cliInstallationResult{err: ErrAuthorizationForbidden}
	}
	if err := checkCLIInstallationReceiptIntent(req, intent, s.now()); err != nil {
		return cliInstallationResult{err: err}
	}
	receipt, err := originalCLIInstallationReceipt(intent, state.cliInstallationLinks[req.key()])
	return cliInstallationResult{receipt: receipt, err: err}
}
