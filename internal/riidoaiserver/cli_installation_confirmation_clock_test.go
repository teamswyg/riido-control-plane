package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestCLIInstallationConfirmationUsesSingleTimeBoundary(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "dynamo"}[durable], func(t *testing.T) {
			var expiry atomic.Pointer[time.Time]
			var samples atomic.Int64
			var armed atomic.Bool
			clock := func() time.Time {
				end := expiry.Load()
				if end == nil || !armed.Load() {
					return time.Now().UTC()
				}
				if samples.Add(1) == 1 {
					return end.Add(-time.Nanosecond)
				}
				return *end
			}
			config := StoreConfig{Now: clock}
			var operations *DynamoDBAssignmentOperationStore
			if durable {
				_, newStore, _ := newTerminationDynamoFixture(t)
				operations = newStore()
				operations.now = clock
				config.OperationStore = operations
			}
			f := newCLIInstallationFixture(t, config)
			intent := f.intent(t)
			expiry.Store(&intent.ExpiresAt)
			if durable {
				armConfirmationClockAfterLinkRead(t, operations, f.request.cliInstallationIdentity, func() { armed.Store(true) })
			} else {
				armed.Store(true)
			}
			path := "/v1/daemon/cli-installation-links/" + intent.IntentID + "/confirm"
			body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}
			confirmed := f.post(t, path, body, true)
			var link cliInstallationLink
			if confirmed.Code != http.StatusOK || json.Unmarshal(confirmed.Body.Bytes(), &link) != nil {
				t.Fatalf("valid boundary confirmation failed: status=%d", confirmed.Code)
			}
			if !link.LinkedAt.Before(intent.ExpiresAt) {
				t.Fatal("confirmation emitted an ACK at or after its expiry boundary")
			}
			receipt := f.post(t, "/v1/internal/cli-installations/link-receipts", installationReceiptBody(f, intent.IntentID), false)
			if receipt.Code != http.StatusOK {
				t.Fatalf("accepted confirmation cannot recover its receipt: status=%d", receipt.Code)
			}
			if replay := f.post(t, path, body, true); replay.Code != http.StatusOK || replay.Body.String() != confirmed.Body.String() {
				t.Fatal("consumed confirmation changed original ACK at expiry")
			}
		})
	}
}
