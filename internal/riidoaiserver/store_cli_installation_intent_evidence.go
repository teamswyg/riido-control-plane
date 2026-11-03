package riidoaiserver

import "context"

func (s *Store) readCLIInstallationIntentEvidence(ctx context.Context, req cliInstallationIntentCreateRequest) (cliInstallationIntentEvidence, error) {
	if s.snapshotStore != nil && s.operationStore != nil {
		return cliInstallationIntentEvidence{}, errCLIInstallationUnavailable
	}
	if durable, ok := s.operationStore.(cliInstallationIntentEvidenceReader); ok {
		return durable.readCLIInstallationIntentEvidence(ctx, req)
	}
	result, err := s.executeCLIInstallation(ctx, cliInstallationCommand{evidence: &req})
	return result.evidence, err
}

func (s *Store) readLocalCLIInstallationIntentEvidence(state *storeState, ctx context.Context, req cliInstallationIntentCreateRequest) cliInstallationResult {
	lookup, err := installationEvidenceLookup(req)
	if err != nil {
		return cliInstallationResult{err: err}
	}
	actor, err := readInstallationEvidence(cliInstallationSnapshot{Intents: state.cliInstallationIntents, IntentReplays: state.cliInstallationReplays}, lookup)
	if err != nil || s.snapshotStore == nil {
		return cliInstallationResult{evidence: actor, err: err}
	}
	// A failed Save may leave an uncommitted pair in actor memory. Read the
	// durable pair without Save, and fail closed on missing/disagreeing evidence.
	snapshot, found, err := s.snapshotStore.LoadStoreSnapshot(ctx)
	if err != nil {
		return cliInstallationResult{err: err}
	}
	if found && snapshot.SchemaVersion != StoreSnapshotSchemaVersion {
		return cliInstallationResult{err: errCLIInstallationUnavailable}
	}
	if !found {
		snapshot = StoreSnapshot{}
	}
	saved, err := readInstallationEvidence(snapshot.CLIInstallations, lookup)
	if err != nil {
		return cliInstallationResult{err: err}
	}
	if saved.Found != actor.Found || (saved.Found && !sameInstallationEvidence(*saved.Intent, *actor.Intent)) {
		return cliInstallationResult{err: errCLIInstallationUnavailable}
	}
	return cliInstallationResult{evidence: saved}
}

func sameInstallationEvidence(a, b cliInstallationIntentReplayResponse) bool {
	return a.SchemaVersion == b.SchemaVersion && a.RequestID == b.RequestID && a.IntentID == b.IntentID && a.Challenge == b.Challenge && a.ExpiresAt.Equal(b.ExpiresAt)
}
