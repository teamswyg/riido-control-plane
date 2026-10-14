package main

import (
	"strings"
	"testing"
)

func TestVerifyExternalEvidenceRejectsBoundaryDrift(t *testing.T) {
	tests := []struct {
		name string
		item externalEvidence
		want string
	}{
		{
			name: "wrong repo",
			item: externalEvidence{
				Risk: requiredRisks[0], Status: "verified", Repo: "riido-control-plane",
				Test: "TestContract", Proves: "contract evidence",
			},
			want: "repo boundary",
		},
		{
			name: "private package path",
			item: externalEvidence{
				Risk: requiredRisks[0], Status: "verified", Repo: "riido-contracts",
				Test: "internal/TestContract", Proves: "contract evidence",
			},
			want: "private package paths",
		},
		{
			name: "missing human doc mention",
			item: externalEvidence{
				Risk: requiredRisks[0], Status: "verified", Repo: "riido-contracts",
				Test: "TestContract", Proves: "contract evidence",
			},
			want: "human doc must mention",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyExternalEvidence(tc.item, "document without the named test")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestVerifyExternalEvidenceRejectsUnknownRiskAndShape(t *testing.T) {
	err := verifyExternalEvidence(externalEvidence{}, "")
	if err == nil || !strings.Contains(err.Error(), "invalid external evidence") {
		t.Fatalf("expected invalid evidence, got %v", err)
	}

	item := externalEvidence{
		Risk: "unknown", Status: "verified", Repo: "riido-contracts",
		Test: "TestContract", Proves: "contract evidence",
	}
	err = verifyExternalEvidence(item, "TestContract")
	if err == nil || !strings.Contains(err.Error(), "unexpected external evidence risk") {
		t.Fatalf("expected unexpected risk, got %v", err)
	}
}
