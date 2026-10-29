package riidoaiserver

import "context"

const cliInstallationIntentReplayItemSchema = "cli_installation_intent_replay/v1"

func cliInstallationIntentReplayKey(key string) map[string]map[string]string {
	return map[string]map[string]string{"pk": {"S": "CLI_INSTALLATION_INTENT_REPLAY#" + key}, "sk": {"S": "REQUEST"}}
}

func (s *DynamoDBAssignmentOperationStore) readCLIInstallationIntentReplay(ctx context.Context, candidate cliInstallationIntentReplay) (cliInstallationIntentReplayResponse, bool, error) {
	key := cliInstallationIntentReplayKey(candidate.key())
	item, err := s.executionTerminationGet(ctx, key["pk"]["S"], key["sk"]["S"])
	if err != nil || len(item) == 0 {
		return cliInstallationIntentReplayResponse{}, false, err
	}
	var original cliInstallationIntentReplay
	if decodeCLIInstallationItem(item, key, cliInstallationIntentReplayItemSchema, &original) != nil || original.valid() != nil || original.key() != candidate.key() {
		return cliInstallationIntentReplayResponse{}, false, errCLIInstallationUnavailable
	}
	intent, _, err := s.loadCLIInstallationIntent(ctx, original.Response.IntentID)
	if err != nil || !original.matches(intent) {
		return cliInstallationIntentReplayResponse{}, false, errCLIInstallationUnavailable
	}
	if original.Request != candidate.Request || original.Response.RequestID != candidate.Response.RequestID {
		return cliInstallationIntentReplayResponse{}, false, errCLIInstallationConflict
	}
	return original.Response, true, nil
}
