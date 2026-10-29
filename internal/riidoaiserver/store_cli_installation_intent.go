package riidoaiserver

import "context"

func (s *Store) createCLIInstallationIntent(ctx context.Context, intent cliInstallationIntent) error {
	if s.snapshotStore != nil && s.operationStore != nil {
		return errCLIInstallationUnavailable
	}
	if durable, ok := s.operationStore.(cliInstallationRepository); ok {
		return durable.createCLIInstallationIntent(ctx, intent)
	}
	_, err := s.executeCLIInstallation(ctx, cliInstallationCommand{create: &intent})
	return err
}
