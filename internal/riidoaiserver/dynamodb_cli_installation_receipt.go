package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) readCLIInstallationReceipt(ctx context.Context, req cliInstallationReceiptRequest) (cliInstallationReceipt, error) {
	intent, _, err := s.loadCLIInstallationIntent(ctx, req.IntentID)
	if err != nil {
		return cliInstallationReceipt{}, err
	}
	if err = checkCLIInstallationReceiptIntent(req, intent, s.now()); err != nil {
		return cliInstallationReceipt{}, err
	}
	// Consumed and immutable link are committed together. Strong reads cannot
	// turn a pending intent into a receipt for another confirmation.
	link, _, err := s.loadCLIInstallationLink(ctx, req.cliInstallationIdentity)
	if err != nil {
		return cliInstallationReceipt{}, errCLIInstallationUnavailable
	}
	return originalCLIInstallationReceipt(intent, link)
}
