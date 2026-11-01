package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationIntentEvidenceRejectsCorruptionWithoutWrites(t *testing.T) {
	for _, broken := range []string{"request-only", "intent-only", "challenge", "expiry", "binding"} {
		t.Run(broken, func(t *testing.T) {
			dynamo, newStore, _ := newTerminationDynamoFixture(t)
			f := newCLIInstallationFixture(t, StoreConfig{OperationStore: newStore()})
			body := installationIntentReplayBody(t, f)
			response := assertInstallationIntentReplayResponse(t, f.post(t, "/v1/internal/cli-installations/link-intents", body, false), body["request_id"])
			record := cliInstallationIntentReplay{f.request, cliInstallationIntentReplayResponse{response, body["request_id"]}}
			record.Request.SchemaVersion = cliInstallationIntentV2Schema
			requestKey := terminationItemKey(cliInstallationIntentReplayKey(record.key()))
			intentKey := terminationItemKey(cliInstallationIntentKey(response.IntentID))
			dynamo.mu.Lock()
			switch broken {
			case "request-only":
				delete(dynamo.items, intentKey)
			case "intent-only":
				delete(dynamo.items, requestKey)
			case "binding":
				intent := record.intent()
				intent.ReplayKey = "changed"
				data, _ := json.Marshal(intent)
				dynamo.items[intentKey]["payload"]["S"] = string(data)
			default:
				if broken == "challenge" {
					record.Response.Challenge = body["request_id"]
				} else {
					record.Response.ExpiresAt = record.Response.ExpiresAt.Add(time.Nanosecond)
				}
				data, _ := json.Marshal(record)
				dynamo.items[requestKey]["payload"]["S"] = string(data)
			}
			before := len(dynamo.requests)
			dynamo.mu.Unlock()
			if got := f.post(t, installationEvidencePath, installationEvidenceBody(body), false); got.Code != http.StatusServiceUnavailable {
				t.Fatalf("corrupt %s status=%d", broken, got.Code)
			}
			assertEvidenceGetsOnly(t, dynamo, before)
		})
	}
}

func assertEvidenceGetsOnly(t *testing.T, dynamo *terminationDynamoFixture, before int) {
	t.Helper()
	dynamo.mu.Lock()
	defer dynamo.mu.Unlock()
	for _, data := range dynamo.requests[before:] {
		var req struct {
			TableName      string
			Key            map[string]map[string]string
			ConsistentRead bool
		}
		if json.Unmarshal(data, &req) != nil || req.TableName == "" || len(req.Key) != 2 || !req.ConsistentRead {
			t.Fatal("evidence invoked something other than an exact consistent GetItem")
		}
	}
}
