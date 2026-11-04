package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestCLIInstallationIntentEvidencePreservesHistoryWithoutRenewal(t *testing.T) {
	for _, durable := range []bool{false, true} {
		for _, state := range []string{"consumed", "expired", "retired"} {
			t.Run(map[bool]string{false: "local", true: "dynamo"}[durable]+"/"+state, func(t *testing.T) {
				var clock atomic.Int64
				clock.Store(time.Now().UTC().UnixNano())
				now := func() time.Time { return time.Unix(0, clock.Load()).UTC() }
				cfg := StoreConfig{Now: now}
				var dynamo *terminationDynamoFixture
				if durable {
					var newStore func() *DynamoDBAssignmentOperationStore
					dynamo, newStore, _ = newTerminationDynamoFixture(t)
					operations := newStore()
					operations.now = now
					cfg.OperationStore = operations
				}
				f := newCLIInstallationFixture(t, cfg)
				body := installationIntentReplayBody(t, f)
				intent := assertInstallationIntentReplayResponse(t, f.post(t, "/v1/internal/cli-installations/link-intents", body, false), body["request_id"])
				confirm := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}
				if state == "consumed" {
					if got := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", confirm, true); got.Code != http.StatusOK {
						t.Fatalf("confirm=%d", got.Code)
					}
				} else {
					clock.Store(intent.ExpiresAt.Add(time.Second).UnixNano())
				}
				if state == "retired" {
					if got := f.post(t, installationRetirePath, installationRetireBody(f, intent.IntentID), false); got.Code != http.StatusOK {
						t.Fatalf("retire=%d", got.Code)
					}
					clock.Store(intent.ExpiresAt.Add(-time.Second).UnixNano())
				}
				before := 0
				if dynamo != nil {
					dynamo.mu.Lock()
					before = len(dynamo.requests)
					dynamo.mu.Unlock()
				}
				got := f.post(t, installationEvidencePath, installationEvidenceBody(body), false)
				var evidence cliInstallationIntentEvidence
				if got.Code != http.StatusOK || json.Unmarshal(got.Body.Bytes(), &evidence) != nil || !evidence.Found || evidence.Intent == nil || evidence.Intent.Challenge != intent.Challenge || !evidence.Intent.ExpiresAt.Equal(intent.ExpiresAt) {
					t.Fatal("history changed immutable original")
				}
				if dynamo != nil {
					assertEvidenceGetsOnly(t, dynamo, before)
				}
				if state != "consumed" {
					if got := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", confirm, true); got.Code != http.StatusGone {
						t.Fatal("evidence renewed expired/retired authority")
					}
				}
			})
		}
	}
}
