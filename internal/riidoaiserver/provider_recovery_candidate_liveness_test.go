package riidoaiserver

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/hostintegration"
)

func TestRecoveryCandidateChecksOnlySelectedProviderBeforeDispatch(t *testing.T) {
	client := NewDevelopmentAIAgentClientStore()
	binding, _ := client.LookupDaemonBinding("agent-owned-codex")
	client.devices[0].DaemonLastSeenAt = time.Now()
	for i := range client.devices[0].Runtimes {
		r := &client.devices[0].Runtimes[i]
		r.DaemonID, r.LastDetectedAt = binding.DaemonID, time.Now()
		r.HealthStatus, r.DiagnosticCode = hostintegration.ProviderHealthHealthy, hostintegration.ProviderDiagnosticNone
	}
	policy := recoveryTestPolicy()
	policy.Fallbacks = append(policy.Fallbacks, assignment.ProviderRecoveryCandidate{Provider: "claude", ModelID: "unavailable-unselected-model"})
	a := Assignment{ID: "asn", AgentID: binding.AgentID, TaskID: "task", RuntimeProvider: "codex", CreatedBy: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID, RecoveryPolicy: policy, State: AssignmentLeased, LeaseToken: "lease", UpdatedAt: time.Now()}
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: client})
	defer store.Close()
	if err := store.recheckRecoveryCandidate(context.Background(), a, "cursor"); err != nil {
		t.Fatalf("unselected candidate blocked healthy dispatch: %v", err)
	}
	client.devices[0].Runtimes[2].HealthStatus = hostintegration.ProviderHealthUnknown
	client.devices[0].Runtimes[2].DiagnosticCode = hostintegration.ProviderDiagnosticAuthProbeFailed
	state := newStoreState()
	req := recoveryEventRequest(a)
	req.Metadata[assignment.RecoveryProviderKey] = "cursor"
	if _, err := store.validateRecoveryEvent(&state, a, req); err == nil {
		t.Fatal("unknown selected candidate permitted a new dispatch")
	}
	for _, phase := range []string{"recovered", "unavailable", "user_action_required", "cancelled"} {
		req.Metadata[assignment.RecoveryStateKey] = phase
		if _, err := store.validateRecoveryEvent(&state, a, req); err != nil {
			t.Fatalf("auth observation erased final %s diagnostic: %v", phase, err)
		}
	}
}
