package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
)

func TestCoverageLifecyclePolicyFencesActualClaimBeforeFinalCAS(t *testing.T) {
	f := newCoverageLifecycleFixture(t, "codex")
	f.link(t)
	before := f.policy(t, "allowed")
	f.assign(t)
	var calls atomic.Int32
	f.dynamo.mu.Lock()
	f.dynamo.beforePolicyFence = func() { calls.Add(1); f.poll(t) }
	f.dynamo.mu.Unlock()
	after := f.policy(t, "blocked")
	f.assertCurrentRevision(t, before, after, calls.Load())
}

func TestCoverageLifecyclePolicyFencesActualRotationBeforeFinalCAS(t *testing.T) {
	f := newCoverageLifecycleFixture(t, "codex")
	f.link(t)
	before := f.policy(t, "allowed")
	original := f.enrollment.CredentialGeneration
	rotated := make(chan EnrollDeviceResponse, 1)
	var calls atomic.Int32
	f.dynamo.mu.Lock()
	f.dynamo.beforePolicyFence = func() {
		calls.Add(1)
		response := f.post(t, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", EnrollDeviceRequest{MachineID: "lifecycle-machine"}, "human", "")
		var credential EnrollDeviceResponse
		if response.Code != http.StatusCreated || json.Unmarshal(response.Body.Bytes(), &credential) != nil || credential.CredentialGeneration == original {
			t.Errorf("interleaved actual rotation failed: status=%d", response.Code)
		}
		rotated <- credential
	}
	f.dynamo.mu.Unlock()
	after := f.policy(t, "allowed")
	select {
	case credential := <-rotated:
		f.enrollment = credential
	default:
		t.Fatal("policy never crossed the rotation hook")
	}
	f.assertCurrentRevision(t, before, after, calls.Load())
}

func (f *coverageLifecycleFixture) assertCurrentRevision(t *testing.T, before, after cliInstallationPolicy, calls int32) {
	t.Helper()
	f.dynamo.mu.Lock()
	current := f.dynamo.items["DEVICE_POLICY#"+f.enrollment.DeviceID+"\x00REVISION"]["revision"]["N"]
	f.dynamo.mu.Unlock()
	old, oldErr := strconv.ParseUint(before.ExecutionRevision, 10, 64)
	next, nextErr := strconv.ParseUint(after.ExecutionRevision, 10, 64)
	if calls != 1 || oldErr != nil || nextErr != nil || next <= old || after.ExecutionRevision != current {
		t.Fatal("policy returned a stale revision after an actual operation crossed its final CAS")
	}
}
