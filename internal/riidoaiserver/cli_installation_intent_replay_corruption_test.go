package riidoaiserver

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationIntentV2RejectsIncompleteOrChangedPairs(t *testing.T) {
	for _, broken := range []string{"request-only", "intent-only", "challenge", "expiry", "binding"} {
		t.Run(broken, func(t *testing.T) {
			dynamo, newStore, _ := newTerminationDynamoFixture(t)
			f := newCLIInstallationFixture(t, StoreConfig{OperationStore: newStore()})
			body := installationIntentReplayBody(t, f)
			first := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
			response := assertInstallationIntentReplayResponse(t, first, body["request_id"])
			record := cliInstallationIntentReplay{f.request, cliInstallationIntentReplayResponse{response, body["request_id"]}}
			record.Request.SchemaVersion = cliInstallationIntentV2Schema
			requestKey := terminationItemKey(cliInstallationIntentReplayKey(record.key()))
			intentKey := terminationItemKey(cliInstallationIntentKey(response.IntentID))
			dynamo.mu.Lock()
			if broken == "request-only" {
				delete(dynamo.items, intentKey)
			} else if broken == "intent-only" {
				delete(dynamo.items, requestKey)
			} else if broken == "binding" {
				intent := record.intent()
				intent.ReplayKey = "changed"
				data, _ := json.Marshal(intent)
				dynamo.items[intentKey]["payload"]["S"] = string(data)
			} else {
				if broken == "challenge" {
					record.Response.Challenge = base64.RawURLEncoding.EncodeToString(make([]byte, 32))
				} else {
					record.Response.ExpiresAt = record.Response.ExpiresAt.Add(time.Nanosecond)
				}
				data, _ := json.Marshal(record)
				dynamo.items[requestKey]["payload"]["S"] = string(data)
			}
			commits := dynamo.commits
			dynamo.mu.Unlock()
			got := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
			if got.Code != http.StatusServiceUnavailable {
				t.Fatalf("broken %s pair recovered or replaced: %d", broken, got.Code)
			}
			dynamo.mu.Lock()
			changed := dynamo.commits != commits
			dynamo.mu.Unlock()
			if changed {
				t.Fatal("broken pair caused replacement publication")
			}
		})
	}
}
