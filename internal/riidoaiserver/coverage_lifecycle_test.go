package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCoverageLifecycleActualEnrollmentClaimReceiptAndRotation(t *testing.T) {
	f := newCoverageLifecycleFixture(t, "codex")
	f.link(t)
	f.policy(t, "allowed")
	f.assign(t)
	for _, body := range []map[string]any{{"wait_ms": 1}, {"device_id": "legacy-forged", "wait_ms": 1}, {"device_id": f.enrollment.DeviceID, "wait_ms": 1}} {
		body["daemon_id"], body["runtime_id"] = f.binding.DaemonID, f.binding.RuntimeID
		response := f.post(t, "/v1/agents/"+f.binding.AgentID+"/poll", body, "bearer", "")
		if body["device_id"] == f.enrollment.DeviceID && response.Code != http.StatusForbidden {
			t.Fatalf("fully matching bearer request bypassed device provenance: status=%d", response.Code)
		}
		if response.Code != http.StatusForbidden && response.Code != http.StatusBadRequest {
			t.Fatalf("raw bearer covered binding status=%d", response.Code)
		}
	}
	f.policy(t, "allowed")
	poll := f.poll(t)
	f.policy(t, "blocked")
	original := f.enrollment
	rotated := f.post(t, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", EnrollDeviceRequest{MachineID: "lifecycle-machine"}, "human", "")
	if rotated.Code != http.StatusCreated || json.Unmarshal(rotated.Body.Bytes(), &f.enrollment) != nil || f.enrollment.ExecutionCoverageGeneration != original.ExecutionCoverageGeneration || f.enrollment.CredentialGeneration == original.CredentialGeneration {
		t.Fatalf("actual active credential rotation status=%d", rotated.Code)
	}
	f.policy(t, "blocked")
	f.terminate(t, poll)
	f.policy(t, "unknown")
	receipt := f.prove(t, poll)
	f.policy(t, "allowed")
	if again := f.prove(t, poll); again != receipt {
		t.Fatal("duplicate proof changed immutable receipt")
	}
	f.policy(t, "allowed")
	f.restart(t)
	f.policy(t, "allowed")
	if again := f.prove(t, poll); again != receipt {
		t.Fatal("restart changed the durable receipt")
	}
	f.dynamo.mu.Lock()
	policy, err := executionCoverageFromItem(f.dynamo.items["DEVICE_POLICY#"+f.enrollment.DeviceID+"\x00REVISION"], f.enrollment.DeviceID)
	f.dynamo.mu.Unlock()
	if err != nil || policy.Unresolved != 0 {
		t.Fatal("proof did not close exactly one unresolved claim")
	}
	t.Log("actual enrollment/link/policy/claim/event/receipt handlers and one atomic synthetic Dynamo table passed; provider observation is synthetic")
}
