package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) readCLIInstallationIntentEvidence(ctx context.Context, req cliInstallationIntentCreateRequest) (cliInstallationIntentEvidence, error) {
	lookup, err := installationEvidenceLookup(req)
	if err != nil {
		return cliInstallationIntentEvidence{}, err
	}
	original, found, err := s.readCLIInstallationIntentReplay(ctx, lookup)
	if err != nil {
		return cliInstallationIntentEvidence{}, err
	}
	if !found {
		// Distinguish an absent pair from a surviving intent without its replay.
		key := cliInstallationIntentKey(lookup.intentID())
		item, err := s.executionTerminationGet(ctx, key["pk"]["S"], key["sk"]["S"])
		if err != nil {
			return cliInstallationIntentEvidence{}, err
		}
		if len(item) != 0 {
			return cliInstallationIntentEvidence{}, errCLIInstallationUnavailable
		}
	}
	return installationEvidenceResponse(lookup, original, found), nil
}
