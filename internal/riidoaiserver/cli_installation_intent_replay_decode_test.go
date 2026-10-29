package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIInstallationIntentV2DiscriminatorPreservesStrictJSON(t *testing.T) {
	f := newCLIInstallationFixture(t)
	body := installationIntentReplayBody(t, f)
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	raw := string(encoded)
	cases := []struct {
		name, body string
		status     int
	}{
		{"duplicate-schema", `{"schema_version":"` + cliInstallationIntentSchema + `",` + raw[1:], http.StatusBadRequest},
		{"duplicate-request", raw[:len(raw)-1] + `,"request_id":"` + body["request_id"] + `"}`, http.StatusBadRequest},
		{"trailing-object", raw + `{}`, http.StatusBadRequest},
		{"unknown-field", raw[:len(raw)-1] + `,"ignored":true}`, http.StatusBadRequest},
		{"oversized", strings.Repeat(" ", 4<<10) + raw, http.StatusRequestEntityTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/internal/cli-installations/link-intents", bytes.NewBufferString(tc.body))
			req.Header.Set("Authorization", "Bearer service-fixture-token")
			response := httptest.NewRecorder()
			f.handler.ServeHTTP(response, req)
			if response.Code != tc.status {
				t.Fatalf("strict body status=%d want=%d", response.Code, tc.status)
			}
		})
	}
	assertInstallationIntentReplayResponse(t, f.post(t, "/v1/internal/cli-installations/link-intents", body, false), body["request_id"])
}
