package main

import (
	"fmt"
	"os"
	"strings"
)

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
	case "outbox":
		return verifyOutboxCase(tc)
	case "snapshot":
		return verifySnapshotCase(tc)
	case "outbox-failure":
		return verifyOutboxFailureCase(tc)
	default:
		return caseEvidence{}, fmt.Errorf("unknown case kind %q", tc.Kind)
	}
}

func verifyExpectedEvents(got, want []string) error {
	if len(got) != len(want) {
		return fmt.Errorf("event count=%d want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			return fmt.Errorf("event[%d]=%s want %s", i, got[i], want[i])
		}
	}
	return nil
}

func verifySourceChecks(root string, checks []sourceCheck) error {
	for _, check := range checks {
		body, err := os.ReadFile(resolve(root, check.File))
		if err != nil {
			return fmt.Errorf("read source check %s: %w", check.Name, err)
		}
		if err := verifyNeedles(check, string(body)); err != nil {
			return err
		}
	}
	return nil
}

func verifyNeedles(check sourceCheck, body string) error {
	for _, needle := range check.Contains {
		if !strings.Contains(body, needle) {
			return fmt.Errorf("source check %s missing %q in %s", check.Name, needle, check.File)
		}
	}
	return nil
}
