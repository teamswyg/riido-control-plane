package contractscompat

import (
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/progressmessage"
	"github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-contracts/task"
)

func TestContractsBaseline(t *testing.T) {
	if !ir.EventTaskQueued.IsTransition() {
		t.Fatal("TaskQueued must remain a transition event")
	}
	if task.FSMSchemaVersion != 1 {
		t.Fatalf("FSMSchemaVersion = %d", task.FSMSchemaVersion)
	}
	if !task.ValidateTransition(task.StateCreated, task.StateQueued, ir.EventTaskQueued) {
		t.Fatal("Created -> Queued must remain legal")
	}
	if task.GeneratedTaskFSMServiceProvider().TaskFSM().Name() != "task" {
		t.Fatal("task FSM service provider must return the generated task FSM")
	}
	if !task.GeneratedTaskFSM().CanTransition(task.TaskStateCodeCreated, task.TaskStateCodeQueued, ir.EventTypeCodeTaskQueued) {
		t.Fatal("Generated task FSM must keep Created -> Queued transition")
	}
	if assignment.SchemaVersion != "riido-ai-server.v1" {
		t.Fatalf("assignment SchemaVersion = %q", assignment.SchemaVersion)
	}
	if !assignment.CanTransition(assignment.AssignmentQueued, assignment.AssignmentLeased) {
		t.Fatal("Queued -> Leased assignment transition must remain legal")
	}
	if assignment.GeneratedAssignmentFSMServiceProvider().AssignmentFSM().Name() != "assignment" {
		t.Fatal("assignment FSM service provider must return the generated assignment FSM")
	}
	if !assignment.GeneratedAssignmentFSM().CanTransition(assignment.AssignmentStateCodeQueued, assignment.AssignmentStateCodeLeased) {
		t.Fatal("Generated assignment FSM must keep queued -> leased transition")
	}
	if assignment.ApprovalTimeoutTerminalStatus != assignment.ApprovalTimedOut {
		t.Fatal("approval timeout must resolve to the timed_out terminal status")
	}
	if !assignment.ApprovalTimedOut.IsTerminal() || assignment.ApprovalPending.IsTerminal() {
		t.Fatal("approval terminal predicates drifted")
	}
	if assignment.ApprovalDecisionApprove.Code() != assignment.ApprovalDecisionCodeApprove {
		t.Fatal("approval decision enum drifted")
	}

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
	rendered, ok := progressmessage.Render(1101, progressmessage.NormalizeArgsForCode(1101, map[string]string{
		"label":       "GitHub 조회 중",
		"description": "이슈 목록",
	}), progressmessage.DefaultLocale)
	if !ok || rendered != "GitHub 수집 중 - 이슈 목록" {
		t.Fatalf("progressmessage.Render = %q, %v", rendered, ok)
	}
}
