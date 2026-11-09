package riidoaiserver

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestCLIInstallationTransferPreservesCredentialsAndFencesSource(t *testing.T) {
	f := newCLITransferFixture(t)
	before, _, err := f.operations.executionCoverageAuthority.LoadAIAgentClientSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	status, link := f.confirm(t)
	if status != http.StatusOK || link.DeviceID != f.target.DeviceID || link.LinkRevision != "2" {
		t.Fatalf("transfer status=%d revision=%s", status, link.LinkRevision)
	}
	after, _, err := f.operations.executionCoverageAuthority.LoadAIAgentClientSnapshot(context.Background())
	if err != nil || !reflect.DeepEqual(before.DeviceCredentials, after.DeviceCredentials) {
		t.Fatal("transfer modified native credentials")
	}
	f.enrollment = f.source
	f.assign(t)
	body := PollRequest{DeviceID: f.source.DeviceID, DaemonID: f.binding.DaemonID, RuntimeID: f.binding.RuntimeID, WaitMs: 1}
	if w := f.post(t, "/v1/agents/"+f.binding.AgentID+"/poll", body, "device", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("retired device poll=%d", w.Code)
	}
	f.cli.request = f.originalRequest
	proof := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.cli.signature(t, f.original)}
	if w := f.post(t, "/v1/daemon/cli-installation-links/"+f.original.IntentID+"/confirm", proof, "device", ""); w.Code != http.StatusConflict {
		t.Fatalf("old proof replay=%d", w.Code)
	}
}
