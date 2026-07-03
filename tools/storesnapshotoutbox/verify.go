package main

import (
	"fmt"
	"os"
)

func verify(root string, m manifest, results []caseEvidence, checkDoc bool) error {
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

func verifyDoc(root string, m manifest) error {
	current, err := os.ReadFile(resolve(root, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(current) != renderDoc(m) {
		return fmt.Errorf("%s is stale; run go run ./tools/storesnapshotoutbox -write-doc", m.GeneratedDoc)
	}
	return nil
}

func verifyCaseNames(cases []caseSpec, results []caseEvidence) error {
	if len(results) != len(cases) {
		return fmt.Errorf("case result count=%d want %d", len(results), len(cases))
	}
	seen := map[string]struct{}{}
	for _, result := range results {
		if result.Name == "" {
			return fmt.Errorf("case evidence has empty name")
		}
		seen[result.Name] = struct{}{}
	}
	for _, tc := range cases {
		if _, ok := seen[tc.Name]; !ok {
			return fmt.Errorf("missing case evidence %s", tc.Name)
		}
	}
	return nil
}
