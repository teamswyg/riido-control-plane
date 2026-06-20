package main

import "fmt"

func verify(root string, m manifest, results []endpointEvidence, checkDoc bool) error {
	if err := verifySourceChecks(root, m.SourceChecks); err != nil {
		return err
	}
	if err := verifyResults(m.Endpoints, results); err != nil {
		return err
	}
	if checkDoc {
		return verifyDoc(root, m)
	}
	return nil
}

func verifyResults(want []endpointContract, got []endpointEvidence) error {
	byName := map[string]endpointEvidence{}
	for _, result := range got {
		byName[result.Name] = result
	}
	for _, endpoint := range want {
		result, ok := byName[endpoint.Name]
		if !ok {
			return fmt.Errorf("missing endpoint evidence %s", endpoint.Name)
		}
		if result.HTTPStatus != endpoint.HTTPStatus || result.Status != endpoint.Status {
			return fmt.Errorf("%s = %d/%s, want %d/%s", endpoint.Name, result.HTTPStatus, result.Status, endpoint.HTTPStatus, endpoint.Status)
		}
	}
	return nil
}
