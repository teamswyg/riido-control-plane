package main

import "fmt"

func verify(root string, m manifest, results []corsEvidence, checkDoc bool) error {
	if err := verifySourceChecks(root, m.SourceChecks); err != nil {
		return err
	}
	if err := verifyCaseNames(m.CORSCases, results); err != nil {
		return err
	}
	if checkDoc {
		return verifyDoc(root, m)
	}
	return nil
}

func verifyCaseNames(cases []corsCase, results []corsEvidence) error {
	if len(results) != len(cases) {
		return fmt.Errorf("cors result count=%d want %d", len(results), len(cases))
	}
	seen := map[string]struct{}{}
	for _, result := range results {
		if result.Name == "" {
			return fmt.Errorf("cors evidence has empty case name")
		}
		seen[result.Name] = struct{}{}
	}
	for _, tc := range cases {
		if _, ok := seen[tc.Name]; !ok {
			return fmt.Errorf("missing cors evidence case %s", tc.Name)
		}
	}
	return nil
}
