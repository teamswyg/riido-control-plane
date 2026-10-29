package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func (f *cliInstallationFixture) post(t *testing.T, path string, body any, device bool) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	if device {
		r.Header.Set(deviceIDHeader, "fixture-device")
		r.Header.Set(deviceSecretHeader, "fixture-device-secret")
	} else {
		r.Header.Set("Authorization", "Bearer service-fixture-token")
	}
	w := httptest.NewRecorder()
	f.handler.ServeHTTP(w, r)
	return w
}

func (f *cliInstallationFixture) intent(t *testing.T) cliInstallationIntentResponse {
	t.Helper()
	w := f.post(t, "/v1/internal/cli-installations/link-intents", f.request, false)
	if w.Code != http.StatusOK {
		t.Fatalf("intent status=%d", w.Code)
	}
	var result cliInstallationIntentResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}
