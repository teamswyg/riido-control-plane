package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-contracts/provider/capability"
)

type StoreSafeRoutingDecision struct {
	Allowed       bool
	ProviderKind  capability.ProviderKind
	RoutingStatus hostintegration.ProviderRoutingStatus
	Reason        string
}

type StoreSafeRoutingInput struct {
	RuntimeProvider  capability.ProviderKind
	ProviderStatuses []ProviderStatusRecord
}

func EvaluateStoreSafeRouting(input StoreSafeRoutingInput) (StoreSafeRoutingDecision, error) {
	provider := capability.ProviderKind(strings.TrimSpace(string(input.RuntimeProvider)))
	if provider == "" {
		return StoreSafeRoutingDecision{}, errors.New("runtime_provider is required")
	}
	decision := StoreSafeRoutingDecision{
		ProviderKind:  provider,
		RoutingStatus: hostintegration.ProviderRoutingUnsupported,
	}
	if len(input.ProviderStatuses) == 0 {
		decision.Allowed = true
		decision.Reason = "provider status not synced"
		return decision, nil
	}
	for _, status := range input.ProviderStatuses {
		if status.ProviderKind != provider {
			continue
		}
		decision.RoutingStatus = status.RoutingStatus
		switch status.RoutingStatus {
		case hostintegration.ProviderRoutingAvailable:
			decision.Allowed = true
			decision.Reason = "provider available"
		case hostintegration.ProviderRoutingLoginRequired:
			decision.Reason = "provider login required"
		case hostintegration.ProviderRoutingUnsupported:
			decision.Reason = "provider unsupported"
		case hostintegration.ProviderRoutingStoreBlocked:
			decision.Reason = "provider blocked by store policy"
		default:
			return StoreSafeRoutingDecision{}, fmt.Errorf("unknown provider routing status %q", status.RoutingStatus)
		}
		return decision, nil
	}
	decision.Reason = "provider status missing"
	return decision, nil
}
