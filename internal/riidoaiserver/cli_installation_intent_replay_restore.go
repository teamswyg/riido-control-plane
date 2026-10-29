package riidoaiserver

func restoreCLIInstallationIntentReplays(state *storeState, snapshot cliInstallationSnapshot) error {
	for key, record := range snapshot.IntentReplays {
		intent, found := snapshot.Intents[record.Response.IntentID]
		if record.valid() != nil || key != record.key() || !found || !record.matches(intent) {
			return errCLIInstallationUnavailable
		}
		state.cliInstallationReplays[key] = record
	}
	for _, intent := range snapshot.Intents {
		if intent.ReplayKey == "" {
			continue // Snapshots written before v2 contain unbound v1 intents.
		}
		record, found := snapshot.IntentReplays[intent.ReplayKey]
		if !found || record.key() != intent.ReplayKey || !record.matches(intent) {
			return errCLIInstallationUnavailable
		}
	}
	return nil
}
