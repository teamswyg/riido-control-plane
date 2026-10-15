package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestHTTPMembershipLegacyAbsentPartFailedPublicationRecovery(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-review-legacy-empty-membership")
	_, _ = newMembershipDynamoReplica(t, s.f)
	// A valid pre-membership split snapshot has all required parts, but neither
	// the optional membership part nor its manifest hash. Opening does not migrate it.
	s.f.mu.Lock()
	for key, item := range s.f.items {
		if dynamoDBAIAgentClientSnapshotPartName(item) == dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships {
			delete(s.f.items, key)
		}
		if item["sk"]["S"] == dynamoDBAIAgentClientSnapshotSK {
			var hashes map[string]string
			if err := json.Unmarshal([]byte(item["part_hashes_json"]["S"]), &hashes); err != nil {
				s.f.mu.Unlock()
				t.Fatal(err)
			}
			delete(hashes, dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships)
			item["part_hashes_json"] = map[string]string{"S": mustJSONDynamoDBAIAgentClientSnapshotPartHashes(hashes)}
		}
	}
	s.f.mu.Unlock()
	client, store := newMembershipDynamoReplica(t, s.f)
	s.f.mu.Lock()
	s.f.rejectScopedPublication = true
	s.f.mu.Unlock()
	posted := s.post(s.replicaHandler(client, store), "/assignment", `{"agent_id":"`+s.agentID+`"}`)
	s.f.mu.Lock()
	s.f.rejectScopedPublication = false
	s.f.mu.Unlock()
	generation, active, err := store.loadTaskAgentDurableExecutions(context.Background(), s.taskID, s.agentID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("POST=%d generation=%s active-after-self-compensation=%d", posted.Code, generation, len(active))
	if posted.Code == http.StatusAccepted {
		t.Fatal("expected publication failure")
	}
	if len(active) != 0 || generation != "" {
		t.Fatal("failed own publication kept a runnable execution and generation in supported legacy snapshot")
	}
}
