package main

import (
	"fmt"
	"sort"
	"strings"
)

const harnessSummaryCheckID = "closed_loop_candidate_harness_summary"

func harnessSummaryProofKey(c check) string {
	return c.Kind + ":" + c.ID
}

func verifyHarnessSummaryCheck(c check, idx indexes) error {
	if c.ID != harnessSummaryCheckID {
		return fmt.Errorf("unknown harness summary check %s", c.ID)
	}
	if len(idx.harnessLoops) == 0 {
		return fmt.Errorf("harness summary must include harness loops")
	}
	for _, loop := range idx.harnessLoops {
		if !contains(loop.PromotesTo, "closed_loop_candidate") {
			return fmt.Errorf("harness %s must promote candidates", loop.ID)
		}
		if len(harnessCandidateArtifacts(loop)) == 0 {
			return fmt.Errorf("harness %s must publish candidate artifact", loop.ID)
		}
	}
	return nil
}

func harnessSummaryProofSurface(idx indexes) *proofSurface {
	return &proofSurface{
		HarnessCount:              len(idx.harnessLoops),
		HarnessLoops:              harnessLoopIDs(idx.harnessLoops),
		HarnessCandidateArtifacts: allHarnessCandidateArtifacts(idx.harnessLoops),
	}
}

func harnessCandidateArtifacts(loop registryLoop) []string {
	out := []string{}
	for _, item := range loop.Evidence {
		if item.Redacted && strings.Contains(item.Path, "closed-loop-candidates") {
			out = append(out, item.Path)
		}
	}
	sort.Strings(out)
	return out
}
