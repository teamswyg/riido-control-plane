package riidoaiserver

import (
	"net/http"
	"testing"
)

// A device enrolled in one workspace may run agents of another workspace; the
// poll path leases those assignments to it, so context reads must follow.
func TestAssignmentContextAllowsBoundDeviceEnrolledInAnotherWorkspace(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	f.agents.mu.Lock()
	record := f.agents.deviceCredentials["device-dev-macbook"]
	record.workspaceID = "workspace-device-home"
	f.agents.deviceCredentials["device-dev-macbook"] = record
	f.agents.mu.Unlock()
	if code, body := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusOK {
		t.Fatalf("bound device from another workspace: status=%d body=%s", code, body)
	}
	if code, _ := f.read(AssignmentContextViewBase, "", "wrong-lease"); code != http.StatusForbidden {
		t.Fatalf("wrong lease must stay forbidden: status=%d", code)
	}
}
