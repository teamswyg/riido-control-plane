package riidoaiserver

import (
	"context"
	"errors"
	"reflect"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/hostintegration"
)

var errRecoveryPolicyChanged = errors.New("provider recovery policy differs from the accepted execution")

func normalizedRecoveryPolicy(policy *assignment.ProviderRecoveryPolicy) (*assignment.ProviderRecoveryPolicy, error) {
	if policy == nil {
		return nil, nil
	}
	normalized, err := policy.Normalized()
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func sameRecoveryPolicy(left, right *assignment.ProviderRecoveryPolicy) bool {
	l, le := left.Normalized()
	r, re := right.Normalized()
	return le == nil && re == nil && reflect.DeepEqual(l, r)
}

func (s Server) authorizeRecoveryPolicy(ctx context.Context, principal AuthorizationResult, selected AgentClientRecord, binding AgentRuntimeBinding, allowExperimental bool, policy *assignment.ProviderRecoveryPolicy) (*assignment.ProviderRecoveryPolicy, error) {
	normalized, err := normalizedRecoveryPolicy(policy)
	if err != nil || normalized == nil || len(normalized.Fallbacks) == 0 {
		return normalized, err
	}
	if selected.OwnerPrincipalID != principal.PrincipalID || binding.DeviceID == "" || binding.DaemonID == "" {
		return nil, ErrAuthorizationForbidden
	}
	inventory, err := s.aiAgent.ListAIAgentDevices(ctx, principal)
	if err != nil {
		return nil, err
	}
	for _, candidate := range normalized.Fallbacks {
		if candidate.Provider == binding.RuntimeProvider {
			return nil, errors.New("fallback must select another provider")
		}
		allowed := false
		for _, device := range inventory.Devices {
			if device.DeviceID != binding.DeviceID || device.OwnerPrincipalID != principal.PrincipalID {
				continue
			}
			for _, runtime := range device.Runtimes {
				if runtime.DeviceID != binding.DeviceID || runtime.OwnerPrincipalID != principal.PrincipalID || runtime.DaemonID != binding.DaemonID || runtimeProviderForAIAgentRuntime(runtime.Kind) != candidate.Provider || !runtimeAvailableForBinding(runtime) || runtime.HealthStatus != hostintegration.ProviderHealthHealthy || runtime.DiagnosticCode != hostintegration.ProviderDiagnosticNone || runtime.RequiresExperimentalOptIn && !allowExperimental {
					continue
				}
				for _, model := range runtime.Models {
					if model.ModelID == candidate.ModelID {
						allowed = true
					}
				}
			}
		}
		if !allowed {
			return nil, errors.New("fallback provider model is not authorized in the selected device inventory")
		}
	}
	return normalized, nil
}
