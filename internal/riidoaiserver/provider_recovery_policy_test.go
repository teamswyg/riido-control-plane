package riidoaiserver

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/hostintegration"
)

func recoveryTestPolicy() *assignment.ProviderRecoveryPolicy {
	return &assignment.ProviderRecoveryPolicy{MaxAttempts: 3, MaxElapsedSeconds: 30, MaxSwitches: 1, Fallbacks: []assignment.ProviderRecoveryCandidate{{Provider: "cursor", ModelID: "cursor-auto"}}}
}

func TestRecoveryPolicyRequiresOwnedSameDeviceHealthyModelInventory(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*DevelopmentAIAgentClientStore, *AgentClientRecord, *AgentRuntimeBinding)
		deny   bool
	}{
		{name: "approved"},
		{name: "other-agent-owner", deny: true, mutate: func(_ *DevelopmentAIAgentClientStore, a *AgentClientRecord, _ *AgentRuntimeBinding) {
			a.OwnerPrincipalID = "other"
		}},
		{name: "other-device", deny: true, mutate: func(_ *DevelopmentAIAgentClientStore, _ *AgentClientRecord, b *AgentRuntimeBinding) {
			b.DeviceID = "other"
		}},
		{name: "unknown-probe", deny: true, mutate: func(s *DevelopmentAIAgentClientStore, _ *AgentClientRecord, _ *AgentRuntimeBinding) {
			s.devices[0].Runtimes[2].HealthStatus = hostintegration.ProviderHealthUnknown
		}},
		{name: "wrong-model", deny: true, mutate: func(s *DevelopmentAIAgentClientStore, _ *AgentClientRecord, _ *AgentRuntimeBinding) {
			s.devices[0].Runtimes[2].Models = nil
		}},
		{name: "wrong-daemon", deny: true, mutate: func(s *DevelopmentAIAgentClientStore, _ *AgentClientRecord, _ *AgentRuntimeBinding) {
			s.devices[0].Runtimes[2].DaemonID = "other"
		}},
		{name: "experimental", deny: true, mutate: func(s *DevelopmentAIAgentClientStore, _ *AgentClientRecord, _ *AgentRuntimeBinding) {
			s.devices[0].Runtimes[2].RequiresExperimentalOptIn = true
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := NewDevelopmentAIAgentClientStore()
			now := time.Now()
			store.devices[0].DaemonLastSeenAt = now
			for i := range store.devices[0].Runtimes {
				r := &store.devices[0].Runtimes[i]
				r.LastDetectedAt = now
				r.DaemonID = "daemon-test"
				r.HealthStatus = hostintegration.ProviderHealthHealthy
				r.DiagnosticCode = hostintegration.ProviderDiagnosticNone
			}
			selected := AgentClientRecord{AgentID: "agent-owned-codex", OwnerPrincipalID: "user-1"}
			binding := AgentRuntimeBinding{AgentID: selected.AgentID, DeviceID: store.devices[0].DeviceID, DaemonID: "daemon-test", RuntimeProvider: "codex"}
			if tc.mutate != nil {
				tc.mutate(store, &selected, &binding)
			}
			policy, err := (Server{aiAgent: store}).authorizeRecoveryPolicy(context.Background(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, selected, binding, false, recoveryTestPolicy())
			if (err != nil) != tc.deny {
				t.Fatalf("policy=%+v error=%v deny=%v", policy, err, tc.deny)
			}
		})
	}
}
