package main

import "fmt"

func candidateFromStaleSource(
	plan dispatchPlan,
	liveStatus string,
	source staleRefreshSource,
	index int,
) closedLoopCandidate {
	return closedLoopCandidate{
		ID:                    fmt.Sprintf("%s:stale-source:%02d", dispatchSourceID, index+1),
		SourceRef:             sourceRefForCandidate(plan, liveStatus),
		Subject:               staleSourceSubject(source),
		HarnessLoop:           dispatchHarnessLoop,
		PromotionTarget:       dispatchPromotionTarget,
		PromotionEdge:         graphEdge{dispatchHarnessLoop, dispatchPromotionTarget, "promotes_failure_to"},
		Observation:           "Loop refresh dispatch could not use stale refresh command source " + source.SourcePath + ".",
		Hypothesis:            "Refresh command evidence must be regenerated before dispatch can safely trigger QA workflows.",
		RequiredNextArtifacts: requiredNextArtifacts(),
		AdoptionPlan:          staleSourceAdoptionPlan(),
	}
}
