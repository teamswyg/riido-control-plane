package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDesktopConnectUsesCredentialDeviceAfterMachineIdentityDrift(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{
		{PrincipalID: "owner", Token: "owner-token", Scopes: []string{"ai-agent:*"}},
		{PrincipalID: "guest", Token: "guest-token", Scopes: []string{"ai-agent:*"}},
	})
	request := func(path, token, body, secret string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set(aiAgentTokenHeader, token)
		if secret != "" {
			req.Header.Set(deviceSecretHeader, secret)
		}
		out := httptest.NewRecorder()
		server.ServeHTTP(out, req)
		return out
	}
	enrolled := request("/v2/desktop/workspaces/owner-workspace/devices/enroll", "owner-token", `{"machine_id":"original-machine"}`, "")
	if enrolled.Code != http.StatusCreated {
		t.Fatal(enrolled.Code, enrolled.Body.String())
	}
	var credential EnrollDeviceResponse
	if err := json.Unmarshal(enrolled.Body.Bytes(), &credential); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]string{"device_id": credential.DeviceID, "machine_id": "changed-machine"})
	for _, token := range []string{"guest-token", "owner-token", "guest-token"} {
		connected := request("/v2/desktop/workspaces/shared-workspace/devices/connect", token, string(payload), credential.DeviceSecret)
		if connected.Code != 200 {
			t.Fatalf("connect status=%d body=%s", connected.Code, connected.Body.String())
		}
		var got struct {
			DeviceID string `json:"device_id"`
		}
		json.Unmarshal(connected.Body.Bytes(), &got)
		if got.DeviceID != credential.DeviceID {
			t.Fatalf("connected ghost device %q; want running device %q", got.DeviceID, credential.DeviceID)
		}
	}
	for _, tc := range []struct{ token, secret string }{{"guest-token", ""}, {"guest-token", "wrong"}, {"", credential.DeviceSecret}} {
		denied := request("/v2/desktop/workspaces/shared-workspace/devices/connect", tc.token, string(payload), tc.secret)
		if denied.Code != http.StatusUnauthorized {
			t.Fatalf("missing/invalid user or device proof status=%d", denied.Code)
		}
	}
}
