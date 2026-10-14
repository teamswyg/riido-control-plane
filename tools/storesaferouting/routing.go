package main

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func verifyRoutingCases(cases []routingCase) ([]routingEvidence, error) {
	results := make([]routingEvidence, 0, len(cases))
	for _, tc := range cases {
		result, err := runRoutingCase(tc)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func runRoutingCase(tc routingCase) (routingEvidence, error) {
	decision, err := riidoaiserver.EvaluateStoreSafeRouting(riidoaiserver.StoreSafeRoutingInput{
		RuntimeProvider:  capability.ProviderKind(tc.RuntimeProvider),
		ProviderStatuses: providerRows(tc.ProviderStatuses),
	})
	result := routingResult(tc, decision, err)
	return result, verifyRoutingResult(tc, result)
}

func providerRows(rows []providerRow) []riidoaiserver.ProviderStatusRecord {
	out := make([]riidoaiserver.ProviderStatusRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, riidoaiserver.ProviderStatusRecord{
			ProviderKind:  capability.ProviderKind(row.ProviderKind),
			RoutingStatus: hostintegration.ProviderRoutingStatus(row.RoutingStatus),
		})
	}
	return out
}

func routingResult(tc routingCase, decision riidoaiserver.StoreSafeRoutingDecision, err error) routingEvidence {
	result := routingEvidence{Name: tc.Name, Runtime: strings.TrimSpace(tc.RuntimeProvider)}
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Allowed = decision.Allowed
	result.RoutingStatus = string(decision.RoutingStatus)
	result.Reason = decision.Reason
	return result
}

func verifyRoutingResult(tc routingCase, got routingEvidence) error {
	if tc.WantErrorContains != "" {
		if !strings.Contains(got.Error, tc.WantErrorContains) {
			return fmt.Errorf("%s error=%q want contains %q", tc.Name, got.Error, tc.WantErrorContains)
		}
		return nil
	}
	if got.Error != "" || got.Allowed != tc.WantAllowed {
		return fmt.Errorf("%s got error=%q allowed=%v", tc.Name, got.Error, got.Allowed)
	}
	if got.RoutingStatus != tc.WantRoutingStatus || got.Reason != tc.WantReason {
		return fmt.Errorf("%s got %s/%q", tc.Name, got.RoutingStatus, got.Reason)
	}
	return nil
}
