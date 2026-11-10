package riidoaiserver

import (
	"maps"
	"net/http"
	"testing"
)

func TestCLIInstallationTransferEvidenceBindsSourceAndScopes(t *testing.T) {
	f := newCLITransferFixture(t)
	original := transferEvidenceBody(t, f)
	for key, value := range map[string]string{"previous_device_id": "other-device-id", "previous_link_revision": "2", "expected_device_id": "other-target-id"} {
		body := maps.Clone(original)
		body[key] = value
		if w := f.post(t, installationEvidencePath, body, "service", ""); w.Code != http.StatusConflict {
			t.Fatalf("changed %s status=%d", key, w.Code)
		}
	}
	for _, scopes := range [][]string{{"cli-installation:link"}, {"cli-installation:transfer"}, {"cli-installation:policy-read"}} {
		service, err := NewCLIInstallationServiceConfig(f.cli.request.Resource, []StaticTokenCredential{{PrincipalID: cliInstallationServicePrincipal, Token: "service-fixture-token", Scopes: scopes}})
		if err != nil {
			t.Fatal(err)
		}
		f.cli.handler = NewServer(ServerConfig{Assignment: f.cli.actor, AIAgentClient: f.devices, CLIInstallationService: service}).Handler()
		if w := f.post(t, installationEvidencePath, original, "service", ""); w.Code != http.StatusForbidden {
			t.Fatalf("incomplete service scopes accepted: status=%d", w.Code)
		}
	}
}

func TestCLIInstallationTransferEvidenceRejectsSchemaMixing(t *testing.T) {
	f := newCLITransferFixture(t)
	for _, change := range []func(map[string]string){
		func(b map[string]string) { b["schema_version"] = cliInstallationIntentEvidenceSchema },
		func(b map[string]string) { delete(b, "previous_device_id") },
		func(b map[string]string) { delete(b, "previous_link_revision") },
		func(b map[string]string) { b["previous_link_revision"] = "01" },
		func(b map[string]string) { b["unknown"] = "private-sentinel" },
	} {
		body := transferEvidenceBody(t, f)
		change(body)
		if w := f.post(t, installationEvidencePath, body, "service", ""); w.Code != http.StatusBadRequest {
			t.Fatalf("invalid transfer evidence status=%d", w.Code)
		}
	}
}
