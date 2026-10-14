package main

import "fmt"

func verifyCases(cases []caseSpec) ([]caseEvidence, error) {
	results := make([]caseEvidence, 0, len(cases))
	for _, tc := range cases {
		result, err := runCase(tc)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func runCase(tc caseSpec) (caseEvidence, error) {
	switch tc.Kind {
	case "provision":
		return verifyProvisionCase(tc)
	case "catalog":
		return verifyCatalogCase(tc)
	case "provider-status":
		return verifyProviderStatusCase(tc)
	case "http":
		return verifyHTTPCase(tc)
	default:
		return caseEvidence{}, fmt.Errorf("unknown review seed case kind %q", tc.Kind)
	}
}

func verifyCaseNames(cases []caseSpec, results []caseEvidence) error {
	if len(results) != len(cases) {
		return fmt.Errorf("case result count=%d want %d", len(results), len(cases))
	}
	seen := map[string]struct{}{}
	for _, result := range results {
		seen[result.Name] = struct{}{}
	}
	for _, tc := range cases {
		if _, ok := seen[tc.Name]; !ok {
			return fmt.Errorf("missing case evidence %s", tc.Name)
		}
	}
	return nil
}
