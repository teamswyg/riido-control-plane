package riidoaiserver

import (
	"context"
)

func (s *PersistentAIAgentClientStore) beginDeviceBootstrap(ctx context.Context, request deviceBootstrapRequest) (deviceBootstrapRecord, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	var empty deviceBootstrapRecord
	if err := request.valid(); err != nil {
		return empty, err
	}
	repo, ok := s.snapshotStore.(deviceBootstrapRepository)
	if !ok {
		return empty, errCLIInstallationUnavailable
	}
	existing, found, err := repo.readDeviceBootstrap(ctx, request.TransitionID)
	if err != nil {
		return empty, err
	}
	if found {
		if existing.Request != request {
			return empty, errCLIInstallationConflict
		}
		return s.currentDeviceBootstrap(ctx, existing, nil)
	}
	if err := s.reloadSnapshot(ctx); err != nil {
		return empty, err
	}
	_, snapshot, err := s.enrollmentTrial()
	if err != nil {
		return empty, err
	}
	before, exists := snapshotDeviceCredential(snapshot, request.ExpectedDeviceID)
	if !exists {
		return empty, ErrAuthorizationForbidden
	}
	intent, _, err := newCLIInstallationIntent(request.cliInstallationIntentRequest)
	if err != nil {
		return empty, err
	}
	generation, err := newCoverageIdentity("cred_")
	if err != nil {
		return empty, err
	}
	after := before
	after.SecretHashSHA256, after.ExecutionCoverageGeneration, after.CredentialGeneration = request.NextSecretHash, bootstrapGeneration(request), generation
	after.IssuedAt = s.snapshotNow().UTC()
	record := deviceBootstrapRecord{Schema: deviceBootstrapSchema, Request: request, Before: before, After: after, Intent: intent, State: "issued", Revision: 1}
	return s.commitDeviceBootstrap(ctx, deviceBootstrapMutation{after: record})
}
