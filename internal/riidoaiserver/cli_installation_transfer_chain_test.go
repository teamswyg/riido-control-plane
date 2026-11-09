package riidoaiserver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCLIInstallationTransferChainRejectsEarlierProofAndReceipt(t *testing.T) {
	f := newCLITransferFixture(t)
	if status, _ := f.confirm(t); status != http.StatusOK {
		t.Fatalf("first transfer=%d", status)
	}
	firstIntent, firstRequest, firstTarget := f.intent, f.cli.request, f.target
	w := f.post(t, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", map[string]string{"machine_id": "third-device", "execution_coverage_protocol": executionCoverageProtocol}, "human", "")
	var third EnrollDeviceResponse
	if w.Code != http.StatusCreated || json.Unmarshal(w.Body.Bytes(), &third) != nil {
		t.Fatalf("third enrollment=%d", w.Code)
	}
	f.cli.request.PreviousDeviceID, f.cli.request.PreviousLinkRevision = firstTarget.DeviceID, "2"
	f.cli.request.ExpectedDeviceID = third.DeviceID
	id, _ := json.Marshal(base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{8}, 32)))
	f.wire = cliInstallationIntentCreateRequest{f.cli.request, id}
	w = f.post(t, "/v1/internal/cli-installations/link-intents", f.wire, "service", "")
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &f.intent) != nil {
		t.Fatalf("second intent=%d", w.Code)
	}
	f.target = third
	if status, link := f.confirm(t); status != http.StatusOK || link.LinkRevision != "3" {
		t.Fatalf("second transfer=%d revision=%s", status, link.LinkRevision)
	}
	f.intent, f.cli.request, f.target = firstIntent, firstRequest, firstTarget
	if status, _ := f.confirm(t); status != http.StatusConflict {
		t.Fatalf("earlier proof replay=%d", status)
	}
	receipt := cliInstallationReceiptRequest{SchemaVersion: cliInstallationReceiptSchema, cliInstallationIdentity: firstRequest.cliInstallationIdentity, IntentID: firstIntent.IntentID, ExpectedDeviceID: firstTarget.DeviceID, ApprovedPrincipalID: firstRequest.ApprovedPrincipalID}
	if w := f.post(t, "/v1/internal/cli-installations/link-receipts", receipt, "service", ""); w.Code == http.StatusOK {
		t.Fatal("earlier receipt restored obsolete link")
	}
}
