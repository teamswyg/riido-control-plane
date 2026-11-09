package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func (f *cliTransferFixture) confirm(t *testing.T) (int, cliInstallationLink) {
	t.Helper()
	f.enrollment = f.target
	body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.cli.signature(t, f.intent.cliInstallationIntentResponse)}
	w := f.post(t, "/v1/daemon/cli-installation-links/"+f.intent.IntentID+"/confirm", body, "device", "")
	var link cliInstallationLink
	if w.Code == http.StatusOK && json.Unmarshal(w.Body.Bytes(), &link) != nil {
		t.Fatal("invalid link")
	}
	return w.Code, link
}

func (f *cliTransferFixture) assertOriginal(t *testing.T) {
	t.Helper()
	link, _, err := f.operations.loadCLIInstallationLink(context.Background(), f.cli.request.cliInstallationIdentity)
	if err != nil || link.DeviceID != f.source.DeviceID || link.LinkRevision != "1" {
		t.Fatal("original link changed")
	}
	c, found, err := f.operations.loadExecutionCoverage(context.Background(), f.source.DeviceID)
	if err != nil || !found || c.State == "revoked" {
		t.Fatal("source unexpectedly retired")
	}
}
