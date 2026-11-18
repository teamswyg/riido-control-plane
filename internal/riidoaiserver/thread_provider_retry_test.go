package riidoaiserver

import (
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
)

func TestCodexPreStartReplyRequiresConclusiveRecovery(t *testing.T) {
	for _, tc := range []struct {
		name, phase, outcome, provider string
		state                          AssignmentState
		seq                            int64
		want                           bool
	}{
		{"preflight exhausted", "unavailable", "not_started", "codex", AssignmentReady, 9, true},
		{"preflight needs action", "user_action_required", "not_started", "codex", AssignmentLeased, 9, true},
		{"intermediate", "checking", "not_started", "codex", AssignmentReady, 9, false},
		{"outcome unknown", "unavailable", "unknown", "codex", AssignmentReady, 9, false},
		{"started", "unavailable", "started", "codex", AssignmentReady, 9, false},
		{"provider changed", "unavailable", "not_started", "claude", AssignmentReady, 9, false},
		{"already running", "unavailable", "not_started", "codex", AssignmentRunning, 9, false},
		{"late receipt", "unavailable", "not_started", "codex", AssignmentReady, 11, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := AssignmentProjection{
				Assignment: Assignment{RuntimeProvider: "codex"}, LastEventSeq: 10,
				Recovery: &recoveryWatermark{Context: assignment.RecoveryContext{State: tc.phase, Provider: tc.provider, OutcomeCertainty: tc.outcome}, Receipts: []recoveryReceipt{{State: tc.state, Seq: tc.seq}}},
			}
			if got := confirmedCodexPreStartFailure(p, &AssignRequest{RuntimeProvider: "codex"}); got != tc.want {
				t.Fatalf("eligible = %v, want %v", got, tc.want)
			}
		})
	}
}
