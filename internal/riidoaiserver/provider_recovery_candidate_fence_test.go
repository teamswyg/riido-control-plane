package riidoaiserver

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func TestRecoveryCandidateFenceRechecksHealthBeforeDispatch(t *testing.T) {
	client := NewDevelopmentAIAgentClientStore()
	binding, ok := client.LookupDaemonBinding("agent-owned-codex")
	if !ok {
		t.Fatal("fixture binding")
	}
	now := time.Now()
	client.devices[0].DaemonLastSeenAt = now
	for i := range client.devices[0].Runtimes {
		runtime := &client.devices[0].Runtimes[i]
		runtime.DaemonID = binding.DaemonID
		runtime.LastDetectedAt = now
		runtime.HealthStatus = hostintegration.ProviderHealthHealthy
		runtime.DiagnosticCode = hostintegration.ProviderDiagnosticNone
	}
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: client})
	defer store.Close()
	current := Assignment{AgentID: binding.AgentID, TaskID: "task", CreatedBy: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID, RecoveryPolicy: recoveryTestPolicy()}
	if err := store.recheckRecoveryCandidate(current, "cursor"); err != nil {
		t.Fatal(err)
	}
	client.devices[0].Runtimes[2].HealthStatus = hostintegration.ProviderHealthUnknown
	client.devices[0].Runtimes[2].DiagnosticCode = hostintegration.ProviderDiagnosticAuthProbeFailed
	if err := store.recheckRecoveryCandidate(current, "cursor"); err == nil {
		t.Fatal("unknown authentication probe authorized fallback dispatch")
	}
}
