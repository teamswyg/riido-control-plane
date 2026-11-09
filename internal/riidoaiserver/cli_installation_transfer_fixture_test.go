package riidoaiserver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
)

type cliTransferFixture struct {
	*coverageLifecycleFixture
	source, target  EnrollDeviceResponse
	original        cliInstallationIntentResponse
	originalRequest cliInstallationIntentRequest
	intent          cliInstallationIntentReplayResponse
	wire            cliInstallationIntentCreateRequest
}

func newCLITransferFixture(t *testing.T, targetWorkspaces ...string) *cliTransferFixture {
	t.Helper()
	f := newCoverageLifecycleFixture(t, "codex")
	original := f.cli.intent(t)
	originalRequest := f.cli.request
	body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.cli.signature(t, original)}
	if w := f.post(t, "/v1/daemon/cli-installation-links/"+original.IntentID+"/confirm", body, "device", ""); w.Code != http.StatusOK {
		t.Fatalf("initial link status=%d", w.Code)
	}
	workspace := "workspace-alpha"
	if len(targetWorkspaces) > 0 {
		workspace = targetWorkspaces[0]
	}
	w := f.post(t, "/v2/desktop/workspaces/"+workspace+"/devices/enroll", map[string]string{"machine_id": "transfer-target", "execution_coverage_protocol": executionCoverageProtocol}, "human", "")
	var target EnrollDeviceResponse
	if w.Code != http.StatusCreated || json.Unmarshal(w.Body.Bytes(), &target) != nil || target.ExecutionCoverageGeneration == "" {
		t.Fatalf("target enrollment status=%d", w.Code)
	}
	service, err := NewCLIInstallationServiceConfig(f.cli.request.Resource, []StaticTokenCredential{{PrincipalID: cliInstallationServicePrincipal, Token: "service-fixture-token", Scopes: []string{"cli-installation:link", "cli-installation:transfer", "cli-installation:policy-read"}}})
	if err != nil {
		t.Fatal(err)
	}
	f.cli.handler = NewServer(ServerConfig{Assignment: f.cli.actor, AIAgentClient: f.devices, Authorizer: f.authorizer, CLIInstallationService: service}).Handler()
	f.cli.request.SchemaVersion = cliInstallationIntentV3Schema
	f.cli.request.PreviousDeviceID, f.cli.request.PreviousLinkRevision = f.enrollment.DeviceID, "1"
	f.cli.request.ExpectedDeviceID = target.DeviceID
	id, _ := json.Marshal(base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)))
	wire := cliInstallationIntentCreateRequest{f.cli.request, id}
	w = f.post(t, "/v1/internal/cli-installations/link-intents", wire, "service", "")
	var intent cliInstallationIntentReplayResponse
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &intent) != nil || intent.SchemaVersion != cliInstallationIntentV3Schema {
		t.Fatalf("transfer intent status=%d", w.Code)
	}
	return &cliTransferFixture{f, f.enrollment, target, original, originalRequest, intent, wire}
}
