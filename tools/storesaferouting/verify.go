package main

import "fmt"

func verify(root string, m manifest, results []routingEvidence, checkDoc bool) error {
	if err := verifySourceChecks(root, m.SourceChecks); err != nil {
		return err
	}
	if err := verifyCaseNames(m.Cases, results); err != nil {
		return err
	}
	if checkDoc {
		return verifyDoc(root, m)
	}
	return nil
}

func verifyCaseNames(cases []routingCase, results []routingEvidence) error {
	if len(results) != len(cases) {
		return fmt.Errorf("routing result count=%d want %d", len(results), len(cases))
	}
	seen := map[string]struct{}{}
	for _, result := range results {
		if result.Name == "" {
			return fmt.Errorf("routing evidence has empty case name")
		}
		seen[result.Name] = struct{}{}
	}
	for _, tc := range cases {
		if _, ok := seen[tc.Name]; !ok {
			return fmt.Errorf("missing routing evidence case %s", tc.Name)
		}
	}
	return nil
}
