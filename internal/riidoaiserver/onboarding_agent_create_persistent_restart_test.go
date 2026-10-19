package riidoaiserver

import (
	"reflect"
	"testing"
)

func onboardingAgentCreatePersistent(t *testing.T, snapshots AIAgentClientSnapshotStore) *PersistentAIAgentClientStore {
	t.Helper()
	store, err := OpenPersistentAIAgentClientStore(t.Context(), NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestOnboardingAgentCreatePersistentSurvivesRestartAndCommitResponseLoss(t *testing.T) {
	for _, loseCommitAck := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal commit", true: "commit then error"}[loseCommitAck], func(t *testing.T) {
			snapshots := &onboardingAgentCreateFaultSnapshot{}
			store := onboardingAgentCreatePersistent(t, snapshots)
			before := len(store.agents)
			snapshots.commitThenError = loseCommitAck
			handler := onboardingAgentCreateHTTPHandler(t, store)
			body := onboardingAgentCreateHTTPBody(t, "durable-response-loss")
			firstRec := postOnboardingAgentCreateHTTP(handler, onboardingAgentCreatePath, body, "ai-agent-token")
			first := decodeOnboardingAgentCreateHTTP(t, firstRec)
			// A separate application instance answers after a lost HTTP response.
			restarted := onboardingAgentCreatePersistent(t, snapshots)
			replayed := decodeOnboardingAgentCreateHTTP(t, postOnboardingAgentCreateHTTP(onboardingAgentCreateHTTPHandler(t, restarted), onboardingAgentCreatePath, body, "ai-agent-token"))
			if !reflect.DeepEqual(first, replayed) || len(restarted.agents) != before+1 {
				t.Fatalf("restart replay changed creation or duplicated the agent: first=%q replay=%q delta=%d", first.Agent.AgentID, replayed.Agent.AgentID, len(restarted.agents)-before)
			}
		})
	}
}
