package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
)

func replayLeaseHandler(t *testing.T, projection, lease map[string]map[string]string, leaseReads, writes *atomic.Int32) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		if r.Header.Get("X-Amz-Target") != dynamoDBGetItemTarget {
			writes.Add(1)
			http.Error(w, "unexpected mutation during live replay", http.StatusBadRequest)
			return
		}
		var request struct {
			Key            map[string]map[string]string `json:"Key"`
			ConsistentRead bool                         `json:"ConsistentRead"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		item := projection
		switch request.Key["pk"]["S"] {
		case "ASSIGNMENT#asn-1":
		case "AGENT#agent-1":
			leaseReads.Add(1)
			if !request.ConsistentRead {
				t.Error("replay lease read is not consistent")
			}
			item = lease
		default:
			t.Errorf("unexpected read key: %v", request.Key)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"Item": item})
	}
}
