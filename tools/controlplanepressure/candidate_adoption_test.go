package main

import (
	"strings"
	"testing"
)

func TestPressureCandidatesCarryClosedLoopAdoptionPath(t *testing.T) {
	report := pressureReport{Findings: []findingEntry{{
		Candidate: candidateForScenario(scenarios()[0]),
	}}}
	assertPressureCandidatesActionable(t, report)
}

func assertPressureCandidatesActionable(t *testing.T, got pressureReport) {
	t.Helper()
	for _, finding := range got.Findings {
		candidate := finding.Candidate
		if candidate.PromotionTarget != pressurePromotionTarget {
			t.Fatalf("finding candidate missing promotion target: %+v", finding)
		}
		if len(candidate.RequiredNextArtifacts) == 0 || len(candidate.AdoptionPlan) == 0 {
			t.Fatalf("finding candidate missing adoption path: %+v", finding)
		}
		assertExecutableAdoptionCommands(t, candidate)
	}
}

func assertExecutableAdoptionCommands(t *testing.T, candidate candidateEntry) {
	t.Helper()
	seen := map[string]bool{}
	for _, step := range candidate.AdoptionPlan {
		seen[step.Artifact] = true
		if !isExecutableAdoptionCommand(step.Command) {
			t.Fatalf("candidate %s has non-executable adoption command: %+v", candidate.ID, step)
		}
	}
	for _, artifact := range candidate.RequiredNextArtifacts {
		if !seen[artifact] {
			t.Fatalf("candidate %s missing adoption command for %s", candidate.ID, artifact)
		}
	}
}

func isExecutableAdoptionCommand(command string) bool {
	return strings.HasPrefix(command, "go run ./tools/") || strings.HasPrefix(command, "go test ./tools/")
}
