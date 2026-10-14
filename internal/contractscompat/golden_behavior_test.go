package contractscompat

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/progressmessage"
	"github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-contracts/task"
)

const contractsCompatGoldenSHA256 = "fdd21cbfff8d56083252b40b5794c3d351cd034626ff695a35cea1deef726093"

func TestContractsCompatBehaviorGolden(t *testing.T) {
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
		t.Fatal(err)
	}
	rendered, _ := progressmessage.Render(1101, progressmessage.NormalizeArgsForCode(1101, map[string]string{
		"label": "GitHub 조회 중", "description": "이슈 목록",
	}), progressmessage.DefaultLocale)
	body, err := json.Marshal(map[string]any{
		"assignment_approval": assignment.ApprovalDecisionApprove.Code(),
		"assignment_schema":   assignment.SchemaVersion,
		"assignment_transition": assignment.GeneratedAssignmentFSM().CanTransition(
			assignment.AssignmentStateCodeQueued, assignment.AssignmentStateCodeLeased),
		"capability_fingerprint": fingerprint,
		"default_model_codex":    providercatalog.DefaultModelID("codex"),
		"progress_1101":          rendered,
		"task_schema":            task.FSMSchemaVersion,
		"task_transition": task.GeneratedTaskFSM().CanTransition(
			task.TaskStateCodeCreated, task.TaskStateCodeQueued, ir.EventTypeCodeTaskQueued),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256Hex(body); got != contractsCompatGoldenSHA256 {
		t.Fatalf("contracts compat golden hash = %s", got)
	}
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
