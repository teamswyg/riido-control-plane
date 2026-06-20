package main

import "fmt"

func verifyProvisionCase(tc caseSpec) (caseEvidence, error) {
	provisioning, err := reviewProvisioning()
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
