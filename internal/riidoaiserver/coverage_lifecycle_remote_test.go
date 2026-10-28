package riidoaiserver

import (
	"testing"
)

func TestCoverageLifecycleRemoteProofDoesNotRestoreAllowed(t *testing.T) {
	f := newCoverageLifecycleFixture(t, "openclaw")
	f.link(t)
	f.assign(t)
	poll := f.poll(t)
	f.policy(t, "blocked")
	f.terminate(t, poll)
	f.policy(t, "unknown")
	f.prove(t, poll)
	f.policy(t, "unknown")
	f.dynamo.mu.Lock()
	policy, err := executionCoverageFromItem(f.dynamo.items["DEVICE_POLICY#"+f.enrollment.DeviceID+"\x00REVISION"], f.enrollment.DeviceID)
	f.dynamo.mu.Unlock()
	if err != nil || policy.Unresolved != 1 {
		t.Fatal("local child proof erased an unobserved remote execution")
	}
}
