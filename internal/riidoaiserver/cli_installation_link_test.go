package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCLIInstallationHTTPLinkAndPolicyNeverInferIdle(t *testing.T) {
	f := newCLIInstallationFixture(t)
	intent := f.intent(t)
	path := "/v1/daemon/cli-installation-links/" + intent.IntentID + "/confirm"
	body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}
	first := f.post(t, path, body, true)
	again := f.post(t, path, body, true)
	if first.Code != http.StatusOK || again.Code != http.StatusOK || first.Body.String() != again.Body.String() {
		t.Fatalf("immutable link status=%d/%d", first.Code, again.Code)
	}
	policyBody := map[string]string{"schema_version": cliInstallationPolicySchema, "installation_id": f.request.InstallationID, "resource": f.request.Resource, "client_id": f.request.ClientID, "key_fingerprint": f.request.KeyFingerprint}
	unknown := f.post(t, "/v1/internal/cli-installations/policy", policyBody, false)
	if unknown.Code != http.StatusOK {
		t.Fatalf("policy=%d", unknown.Code)
	}
	var policy cliInstallationPolicy
	if err := json.Unmarshal(unknown.Body.Bytes(), &policy); err != nil {
		t.Fatal(err)
	}
	if policy.Decision != "unknown" || policy.ExecutionRevision != "1" || policy.CoverageState != "unknown" {
		t.Fatalf("empty history became permission: %+v", policy)
	}
	_, err := f.actor.AssignTask(t.Context(), "task", AssignRequest{ComponentID: "component", AgentID: "agent", RuntimeProvider: "codex", Prompt: "fixture", CreatedBy: "requester"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(t.Context(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{"fixture-device", "owner"})
	if _, err = f.actor.PollAgent(ctx, "agent", PollRequest{}); err != nil {
		t.Fatal(err)
	}
	blocked := f.post(t, "/v1/internal/cli-installations/policy", policyBody, false)
	if blocked.Code != http.StatusOK || json.Unmarshal(blocked.Body.Bytes(), &policy) != nil || policy.Decision != "blocked" || policy.ExecutionRevision != "2" {
		t.Fatalf("claim policy status=%d verdict=%+v", blocked.Code, policy)
	}
}
