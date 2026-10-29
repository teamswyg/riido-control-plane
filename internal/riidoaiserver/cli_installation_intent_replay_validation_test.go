package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestCLIInstallationIntentV2RequiresCanonicalRequestIdentity(t *testing.T) {
	f := newCLIInstallationFixture(t)
	for _, value := range []any{nil, "", "short", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", 123} {
		body := map[string]any{}
		for key, entry := range installationIntentReplayBody(t, f) {
			body[key] = entry
		}
		body["request_id"] = value
		if got := f.post(t, "/v1/internal/cli-installations/link-intents", body, false); got.Code != http.StatusBadRequest {
			t.Fatalf("invalid request identity accepted: %d", got.Code)
		}
		body["schema_version"] = cliInstallationIntentSchema
		if got := f.post(t, "/v1/internal/cli-installations/link-intents", body, false); got.Code != http.StatusBadRequest {
			t.Fatalf("v1 accepted explicit replay field: %d", got.Code)
		}
	}
	body := installationIntentReplayBody(t, f)
	delete(body, "request_id")
	if got := f.post(t, "/v1/internal/cli-installations/link-intents", body, false); got.Code != http.StatusBadRequest {
		t.Fatalf("missing request identity accepted: %d", got.Code)
	}
}

func TestCLIInstallationIntentV2DerivationSeparatesDomainAndTuple(t *testing.T) {
	baseline := installationIntentTupleDigest("domain", "ab", "c")
	if baseline == installationIntentTupleDigest("domain", "a", "bc") || baseline == installationIntentTupleDigest("other-domain", "ab", "c") {
		t.Fatal("tuple framing or domain separation collided")
	}
	f := newCLIInstallationFixture(t)
	body := installationIntentReplayBody(t, f)
	record := cliInstallationIntentReplay{Request: f.request, Response: cliInstallationIntentReplayResponse{RequestID: body["request_id"]}}
	original := record.intentID()
	record.Request.Resource = "https://another.fixture.invalid"
	if record.intentID() == original {
		t.Fatal("different canonical resource shared an intent ID")
	}
	if installationIntentRequestDigest("riido-cli-installation-intent-id.v2", f.request, body["request_id"]) == installationIntentRequestDigest("riido-cli-installation-intent-request.v1", f.request, body["request_id"]) {
		t.Fatal("request lookup and public intent reused one domain")
	}
}
