package riidoaiserver

import (
	"testing"
	"time"
)

func TestCLIInstallationLocalPolicyRequiresFreshLease(t *testing.T) {
	now := time.Date(2026, 10, 28, 1, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name    string
		updated time.Time
		want    string
	}{
		{"fresh", now.Add(-time.Minute + time.Nanosecond), "blocked"},
		{"exact expiry", now.Add(-time.Minute), "unknown"},
		{"expired", now.Add(-time.Minute - time.Nanosecond), "unknown"},
		{"missing timestamp", time.Time{}, "unknown"},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := &Store{now: func() time.Time { return now }, activeLeaseDuration: time.Minute}
			state := newStoreState()
			id := cliInstallationIdentity{InstallationID: "installation"}
			state.cliInstallationLinks[id.key()] = cliInstallationLink{cliInstallationIdentity: id, DeviceID: "device", DeviceOwnerAtLink: "owner", LinkApprovedBy: "owner", LinkRevision: "1"}
			state.devicePolicyRevisions["device"] = 1
			a := Assignment{ID: "assignment", AgentID: "agent", State: AssignmentRunning, LeaseToken: "fixture-lease", UpdatedAt: test.updated}
			state.assignments[a.ID] = a
			state.executionClaims["claim"] = executionClaim{DeviceID: "device", OwnerPrincipalID: "owner", RequesterPrincipalID: "other", AssignmentID: a.ID, AgentID: a.AgentID, LeaseEpoch: leaseEpoch(a.LeaseToken)}
			result := s.readLocalCLIInstallation(&state, cliInstallationRead{identity: id})
			policy := installationPolicyFromRead(result.read, "owner", now)
			if result.err != nil || policy.Decision != test.want || policy.CoverageState != "unknown" {
				t.Fatalf("policy=%+v err=%v want=%s", policy, result.err, test.want)
			}
		})
	}
}
