package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/reviewaccountseed/seedruntime"
)

func verifyProvisionCase(tc caseSpec) (caseEvidence, error) {
	provisioning, err := seedruntime.ReviewProvisioning()
	if err != nil {
		return caseEvidence{}, err
	}
	result := caseEvidence{
		Name:             tc.Name,
		Kind:             tc.Kind,
		TokenHashPresent: provisioning.Credential.TokenSHA256 != "",
		RawTokenPresent:  provisioning.Credential.Token != "",
	}
	if result.TokenHashPresent != tc.WantTokenHashPresent ||
		result.RawTokenPresent != tc.WantRawTokenPresent {
		return result, fmt.Errorf("%s result=%+v", tc.Name, result)
	}
	return result, nil
}
