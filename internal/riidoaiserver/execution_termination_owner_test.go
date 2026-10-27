package riidoaiserver

import (
	"crypto/sha256"
	"testing"
)

func TestExecutionTerminationOwnerAndDeviceAreClaimBound(t *testing.T) {
	for _, changed := range []string{"owner", "device"} {
		t.Run(changed, func(t *testing.T) {
			f := newTerminationHTTPFixture(t, StoreConfig{})
			r := f.request(terminationFixtureBody)
			if changed == "owner" {
				c := f.devices.deviceCredentials["device"]
				c.ownerPrincipalID = "new-owner"
				f.devices.deviceCredentials["device"] = c
			} else {
				f.devices.deviceCredentials["other"] = deviceCredentialRecord{deviceID: "other", secretHash: sha256.Sum256([]byte("secret-sentinel")), ownerPrincipalID: "owner"}
				r.Header.Set(deviceIDHeader, "other")
			}
			w := serveTermination(f.handler, r)
			if w.Code != 403 {
				t.Fatalf("status=%d", w.Code)
			}
		})
	}
}
