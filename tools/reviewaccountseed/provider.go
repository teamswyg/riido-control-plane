package main

import (
	"fmt"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func verifyProviderStatusCase(tc caseSpec) (caseEvidence, error) {
	provisioning, err := reviewProvisioning()
	if err != nil {
		return caseEvidence{}, err
	}
	available := 0
	for _, provider := range provisioning.SyntheticProviderStatusRequest.Providers {
		if provider.RoutingStatus == hostintegration.ProviderRoutingAvailable {
			available++
		}
	}
	result := caseEvidence{
		Name: tc.Name, Kind: tc.Kind,
		ProviderCount:  len(provisioning.SyntheticProviderStatusRequest.Providers),
		AvailableCount: available,
		Channel:        string(provisioning.SyntheticProviderStatusRequest.DistributionChannel),
	}
	if result.ProviderCount != tc.WantProviderCount ||
		result.AvailableCount != tc.WantAvailableCount ||
		result.Channel != tc.WantChannel {
		return result, fmt.Errorf("%s result=%+v", tc.Name, result)
	}
	return result, nil
}
