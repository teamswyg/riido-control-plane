package riidoaiserver

import "testing"

func assertIntentDialoguePollWork(t *testing.T, store *Store, action AIAgentTaskActionResponse) {
	t.Helper()
	pollWork, err := store.PollAgent(t.Context(), action.AgentID, intentDialoguePollRequest())
	if err != nil {
		t.Fatalf("PollAgent after followup: %v", err)
	}
	if pollWork.Assignment == nil || (pollWork.Action != PollStart && pollWork.Action != PollActive) {
		t.Fatalf("followup assignment was not durable: %+v", pollWork)
	}
}

func intentDialoguePollRequest() PollRequest {
	return PollRequest{
		DaemonID:  "daemon-dev-macbook",
		DeviceID:  "device-dev-macbook",
		RuntimeID: "runtime-codex-dev",
	}
}
