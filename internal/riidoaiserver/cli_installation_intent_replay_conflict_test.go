package riidoaiserver

import (
	"maps"
	"net/http"
	"testing"
)

func TestCLIInstallationIntentV2RejectsChangedImmutablePayload(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "dynamo"}[durable], func(t *testing.T) {
			var config StoreConfig
			if durable {
				_, newStore, _ := newTerminationDynamoFixture(t)
				config.OperationStore = newStore()
			}
			f := newCLIInstallationFixture(t, config)
			body := installationIntentReplayBody(t, f)
			first := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
			assertInstallationIntentReplayResponse(t, first, body["request_id"])
			for _, field := range []string{"installation_id", "expected_device_id", "approved_principal_id"} {
				changed := maps.Clone(body)
				changed[field] = "changed"
				if got := f.post(t, "/v1/internal/cli-installations/link-intents", changed, false); got.Code != http.StatusConflict {
					t.Fatalf("changed %s accepted: %d", field, got.Code)
				}
			}
			replacement := newCLIInstallationFixture(t)
			changed := maps.Clone(body)
			changed["public_key"], changed["key_fingerprint"] = replacement.request.PublicKey, replacement.request.KeyFingerprint
			if got := f.post(t, "/v1/internal/cli-installations/link-intents", changed, false); got.Code != http.StatusConflict {
				t.Fatalf("replacement key accepted: %d", got.Code)
			}
			if again := f.post(t, "/v1/internal/cli-installations/link-intents", body, false); again.Body.String() != first.Body.String() {
				t.Fatal("conflicting request replaced original")
			}
		})
	}
}
