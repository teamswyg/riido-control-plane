package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestCLIInstallationIntentV2ReplaysOriginalResponse(t *testing.T) {
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
			intent := assertInstallationIntentReplayResponse(t, first, body["request_id"])
			second := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
			if second.Code != http.StatusOK || second.Body.String() != first.Body.String() {
				t.Fatal("same request did not return its original exact response")
			}
			confirm := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}
			if got := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", confirm, true); got.Code != http.StatusOK {
				t.Fatalf("replayed intent could not be confirmed: %d", got.Code)
			}
			consumed := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
			if consumed.Code != http.StatusOK || consumed.Body.String() != first.Body.String() {
				t.Fatal("consumption changed the original generation response")
			}
		})
	}
}

func TestCLIInstallationIntentV1RejectsReplayRequestID(t *testing.T) {
	f := newCLIInstallationFixture(t)
	body := installationIntentReplayBody(t, f)
	body["schema_version"] = cliInstallationIntentSchema
	if got := f.post(t, "/v1/internal/cli-installations/link-intents", body, false); got.Code != http.StatusBadRequest {
		t.Fatalf("v1 implied unsupported replay authority: %d", got.Code)
	}
	f.intent(t)
}
