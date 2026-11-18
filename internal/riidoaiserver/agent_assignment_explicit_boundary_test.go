package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestExplicitAssignmentReadRefusesUnsafeBoundaries(t *testing.T) {
	for _, name := range []string{"denied", "identity", "ended", "credential", "audit", "requester"} {
		t.Run(name, func(t *testing.T) {
			f := explicitReadFixture(t, func(f *assignmentContextFixture, v *explicitAssignmentResult, w http.ResponseWriter, r *http.Request) bool {
				switch name {
				case "denied":
					w.WriteHeader(http.StatusForbidden)
					return false
				case "identity":
					v.PrincipalID = "different-person"
				case "ended":
					f.event(AssignmentReady, EventAssignmentReady)
					f.event(AssignmentRunning, EventAssignmentRunning)
					f.event(AssignmentCompleted, EventAssignmentCompleted)
				case "credential":
					f.agents.mu.Lock()
					delete(f.agents.deviceCredentials, "device-dev-macbook")
					f.agents.mu.Unlock()
				case "requester":
					t.Error("mismatched requester must not reach source")
				}
				return true
			})
			if name == "audit" {
				f.service.config.AssignmentContext = cacheAuditFailureStore{f.contexts}
				f.server = f.service.Handler()
			}
			if name == "requester" {
				f.contexts.mu.Lock()
				v := f.contexts.snapshots[f.leased.ID]
				v.RequesterPrincipalID = "different-person"
				f.contexts.snapshots[f.leased.ID] = v
				f.contexts.mu.Unlock()
			}
			code, body := f.read("activity", "target-one", f.leased.LeaseToken)
			if code == 200 || strings.Contains(body, "readable evidence") || strings.Contains(body, "Current title") {
				t.Fatalf("boundary leaked status=%d body=%s", code, body)
			}
		})
	}
}
