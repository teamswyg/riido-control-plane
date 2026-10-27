package riidoaiserver

import (
	"strconv"
	"testing"
)

func TestCLIInstallationDynamoPolicyRequiresCurrentLease(t *testing.T) {
	for _, name := range []string{"fresh", "missing", "expired", "exact expiry", "other assignment", "other epoch", "empty epoch", "invalid expiry"} {
		t.Run(name, func(t *testing.T) {
			f, store, dynamo, active := newCLIInstallationLeaseFixture(t)
			want := "unknown"
			switch name {
			case "fresh":
				want = "blocked"
			case "expired":
				active["lease_expires_unix_ms"]["N"] = strconv.FormatInt(dynamo.now.UnixMilli()-1, 10)
			case "exact expiry":
				active["lease_expires_unix_ms"]["N"] = strconv.FormatInt(dynamo.now.UnixMilli(), 10)
			case "other assignment":
				active["active_assignment_id"]["S"] = "other"
			case "other epoch":
				active["lease_token"]["S"] = "other-lease"
			case "empty epoch":
				active["lease_token"]["S"] = ""
			case "invalid expiry":
				active["lease_expires_unix_ms"]["N"] = "invalid"
			}
			dynamo.mu.Lock()
			if name != "missing" {
				dynamo.items[terminationItemKey(active)] = active
			}
			dynamo.mu.Unlock()
			policy, err := store.readCLIInstallationPolicy(t.Context(), f.request.cliInstallationIdentity, f.devices)
			if name == "invalid expiry" {
				if err == nil {
					t.Fatal("invalid active lease must fail closed")
				}
				return
			}
			if err != nil || policy.Decision != want || policy.ExecutionRevision != "2" || policy.CoverageState != "unknown" {
				t.Fatalf("policy=%+v err=%v want=%s", policy, err, want)
			}
		})
	}
}
