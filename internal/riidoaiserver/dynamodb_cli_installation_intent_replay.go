package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) createOrReplayCLIInstallationIntent(ctx context.Context, candidate cliInstallationIntentReplay) (cliInstallationIntentReplayResponse, error) {
	if err := candidate.valid(); err != nil {
		return cliInstallationIntentReplayResponse{}, err
	}
	for range 3 {
		if err := ctx.Err(); err != nil {
			return cliInstallationIntentReplayResponse{}, err
		}
		response, found, err := s.readCLIInstallationIntentReplay(ctx, candidate)
		if err != nil || found {
			return response, err
		}
		actions, err := s.cliInstallationIntentReplayActions(candidate)
		if err != nil {
			return cliInstallationIntentReplayResponse{}, err
		}
		_, err = s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
		if err == nil {
			return candidate.Response, nil
		}
	}
	response, found, err := s.readCLIInstallationIntentReplay(ctx, candidate)
	if err != nil || found {
		return response, err
	}
	return cliInstallationIntentReplayResponse{}, errCLIInstallationUnavailable
}

func (s *DynamoDBAssignmentOperationStore) cliInstallationIntentReplayActions(candidate cliInstallationIntentReplay) ([]dynamoDBTransactWriteItem, error) {
	intent := candidate.intent()
	intentItem, err := cliInstallationItem(cliInstallationIntentKey(intent.ID), cliInstallationIntentItemSchema, intent)
	if err != nil {
		return nil, err
	}
	replayItem, err := cliInstallationItem(cliInstallationIntentReplayKey(candidate.key()), cliInstallationIntentReplayItemSchema, candidate)
	if err != nil {
		return nil, err
	}
	return []dynamoDBTransactWriteItem{s.installationItemPut(intentItem, nil), s.installationItemPut(replayItem, nil)}, nil
}
