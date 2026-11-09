package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationTransferRejectsUnsafeSourceWithoutChangingLink(t *testing.T) {
	for _, mutation := range []string{"active", "unknown", "owner", "revision", "expired", "authority", "workspace"} {
		t.Run(mutation, func(t *testing.T) {
			f := newCLITransferFixture(t, map[bool]string{true: "workspace-other", false: "workspace-alpha"}[mutation == "workspace"])
			if mutation == "active" {
				f.assign(t)
				f.poll(t)
			}
			if mutation == "unknown" || mutation == "owner" {
				f.dynamo.mu.Lock()
				item := f.dynamo.items[membershipDynamoKey(devicePolicyKey(f.source.DeviceID))]
				if mutation == "unknown" {
					item["coverage_state"] = map[string]string{"S": "unknown"}
				} else {
					item["owner_principal_id"] = map[string]string{"S": "other-owner"}
				}
				f.dynamo.mu.Unlock()
			}
			if mutation == "revision" {
				f.cli.request.PreviousLinkRevision = "2"
			}
			if mutation == "expired" {
				f.operations.now = func() time.Time { return f.intent.ExpiresAt }
			}
			if mutation == "authority" {
				f.operations.executionCoverageAuthority = nil
			}
			status, _ := f.confirm(t)
			if status == http.StatusOK {
				t.Fatal("unsafe transfer succeeded")
			}
			f.assertOriginal(t)
		})
	}
}
