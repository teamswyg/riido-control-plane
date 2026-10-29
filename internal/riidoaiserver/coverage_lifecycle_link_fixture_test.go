package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func (f *coverageLifecycleFixture) link(t *testing.T) {
	t.Helper()
	intent := f.cli.intent(t)
	body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.cli.signature(t, intent)}
	if response := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", body, "device", ""); response.Code != http.StatusOK {
		t.Fatalf("actual link confirm status=%d", response.Code)
	}
}

func (f *coverageLifecycleFixture) policy(t *testing.T, want string) cliInstallationPolicy {
	t.Helper()
	body := struct {
		SchemaVersion string `json:"schema_version"`
		cliInstallationIdentity
	}{cliInstallationPolicySchema, f.cli.request.cliInstallationIdentity}
	response := f.post(t, "/v1/internal/cli-installations/policy", body, "service", "")
	var policy cliInstallationPolicy
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &policy) != nil {
		t.Fatalf("actual policy status=%d", response.Code)
	}
	if policy.Decision != want {
		t.Fatalf("policy decision=%s want=%s", policy.Decision, want)
	}
	if want == "allowed" && (policy.CoverageID != f.enrollment.ExecutionCoverageGeneration || policy.CoverageState != "complete") {
		t.Fatal("allowed response lacks exact complete generation")
	}
	return policy
}
