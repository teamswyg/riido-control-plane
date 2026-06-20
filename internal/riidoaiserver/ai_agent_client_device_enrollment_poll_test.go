package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
)

func (f *deviceEnrollmentHTTPFixture) verifyDeviceCredentialPollAuthorization() {
	t := f.t
	body := `{"daemon_id":"daemon-enrolled","device_id":"` + f.enrollment.DeviceID + `","runtime_id":"runtime-codex-dev"}`
	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(body))
	pollReq.Header.Set(deviceIDHeader, f.enrollment.DeviceID)
	pollReq.Header.Set(deviceSecretHeader, f.enrollment.DeviceSecret)
	pollResp := httptest.NewRecorder()
	f.server.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusOK {
		t.Fatalf("poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}
	badPollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(body))
	badPollReq.Header.Set(deviceIDHeader, f.enrollment.DeviceID)
	badPollReq.Header.Set(deviceSecretHeader, "wrong-secret")
	badPollResp := httptest.NewRecorder()
	f.server.ServeHTTP(badPollResp, badPollReq)
	if badPollResp.Code != http.StatusUnauthorized {
		t.Fatalf("bad poll status=%d body=%s", badPollResp.Code, badPollResp.Body.String())
	}
}
