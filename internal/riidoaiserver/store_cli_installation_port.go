package riidoaiserver

import "context"

type cliInstallationConfirmation struct {
	id        string
	device    authenticatedClaimDevice
	signature string
	resource  string
}
type cliInstallationRead struct {
	identity cliInstallationIdentity
	expected *cliInstallationReadSet
}
type cliInstallationReadSet struct {
	link              cliInstallationLink
	revision          uint64
	found             bool
	observation       deviceExecutionObservation
	coverage          *executionCoverage
	coverageAuthority bool
}
type cliInstallationCommand struct {
	ctx     context.Context
	create  *cliInstallationIntent
	replay  *cliInstallationIntentReplay
	confirm *cliInstallationConfirmation
	read    *cliInstallationRead
	receipt *cliInstallationReceiptRequest
	retire  *cliInstallationRetirementRequest
	reply   chan cliInstallationResult
}
type cliInstallationResult struct {
	link    cliInstallationLink
	read    cliInstallationReadSet
	receipt cliInstallationReceipt
	replay  cliInstallationIntentReplayResponse
	retired cliInstallationRetirement
	err     error
}

func (s *Store) executeCLIInstallation(ctx context.Context, command cliInstallationCommand) (cliInstallationResult, error) {
	if s.operationStore != nil {
		return cliInstallationResult{}, errCLIInstallationUnavailable
	}
	command.ctx, command.reply = ctx, make(chan cliInstallationResult, 1)
	if err := s.send(ctx, command); err != nil {
		return cliInstallationResult{}, err
	}
	select {
	case result := <-command.reply:
		return result, result.err
	case <-ctx.Done():
		return cliInstallationResult{}, ctx.Err()
	}
}

func (s *Store) confirmCLIInstallationLink(ctx context.Context, id string, device authenticatedClaimDevice, signature, resource string) (cliInstallationLink, error) {
	if s.snapshotStore != nil && s.operationStore != nil {
		return cliInstallationLink{}, errCLIInstallationUnavailable
	}
	if durable, ok := s.operationStore.(cliInstallationRepository); ok {
		return durable.confirmCLIInstallationLink(ctx, id, device, signature, resource)
	}
	result, err := s.executeCLIInstallation(ctx, cliInstallationCommand{confirm: &cliInstallationConfirmation{id, device, signature, resource}})
	return result.link, err
}
