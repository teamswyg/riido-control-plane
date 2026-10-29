package riidoaiserver

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const installationIntentV2TestSchema = "riido-cli-installation-link-intent.v2"

func installationIntentReplayBody(t *testing.T, f *cliInstallationFixture) map[string]string {
	t.Helper()
	encoded, err := json.Marshal(f.request)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]string
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatal(err)
	}
	body["schema_version"] = installationIntentV2TestSchema
	body["request_id"] = base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	return body
}

func assertInstallationIntentReplayResponse(t *testing.T, response *httptest.ResponseRecorder, requestID string) cliInstallationIntentResponse {
	t.Helper()
	var decoded struct {
		cliInstallationIntentResponse
		RequestID string `json:"request_id"`
	}
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &decoded) != nil {
		t.Fatalf("intent response status=%d", response.Code)
	}
	if decoded.SchemaVersion != installationIntentV2TestSchema || decoded.RequestID != requestID || decoded.IntentID == "" || decoded.Challenge == "" || decoded.ExpiresAt.IsZero() {
		t.Fatal("intent response lost its original request identity")
	}
	return decoded.cliInstallationIntentResponse
}
