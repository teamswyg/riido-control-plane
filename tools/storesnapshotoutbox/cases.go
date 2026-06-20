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
