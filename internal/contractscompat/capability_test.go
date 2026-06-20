package contractscompat

import (
	"testing"

	"github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func TestCapabilityFingerprintBaseline(t *testing.T) {
	fingerprint, err := capability.ComputeCapabilityFingerprint(capability.CapabilityFingerprintInput{
		ProviderKind:          capability.ProviderKind("codex"),
		ProtocolKind:          capability.ProtocolCodexAppServer,
		ProviderVersion:       "codex test",
		DetectedFingerprint:   capability.DetectedFingerprint("detected"),
		AdapterID:             "codex",
		AdapterVersion:        "riido-control-plane-compat.v1",
		ProtocolVersion:       "v1",
		DefaultSandboxMode:    "workspace-write",
		DefaultApprovalPolicy: "on-request",
		PolicyBundleVersion:   "policy-bundle.test.v1",
		ImportantSurfaceFlags: map[string]any{"assignmentRouting": true},
	})
	if err != nil {
		t.Fatalf("ComputeCapabilityFingerprint: %v", err)
	}
	if fingerprint == "" {
		t.Fatal("CapabilityFingerprint is empty")
	}
	if providercatalog.DefaultModelID("codex") != "codex-default" {
		t.Fatal("codex default model must remain stable for assignment projections")
	}
}
