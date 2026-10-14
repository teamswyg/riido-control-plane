package main

import "fmt"

func candidateFromIgnoredCommand(
	plan dispatchPlan,
	liveStatus string,
	command selectedRefreshCommand,
	index int,
) closedLoopCandidate {
	return closedLoopCandidate{
		ID:              fmt.Sprintf("%s:%02d:%s:%s", dispatchSourceID, index+1, command.LoopID, command.Kind),
		SourceRef:       sourceRefForCandidate(plan, liveStatus),
		Subject:         ignoredCommandSubject(command),
		HarnessLoop:     dispatchHarnessLoop,
		PromotionTarget: dispatchPromotionTarget,
		PromotionEdge:   graphEdge{dispatchHarnessLoop, dispatchPromotionTarget, "promotes_failure_to"},
		Observation: "Loop refresh dispatch could not execute " + command.Kind +
			" for loop " + command.LoopID + ".",
		Hypothesis: "The ignored command must become an explicit adoption step so " +
			"refresh automation does not depend on human memory.",
		RequiredNextArtifacts: requiredNextArtifacts(),
		AdoptionPlan:          adoptionPlan(command),
	}
}

func sourceRefForCandidate(plan dispatchPlan, liveStatus string) candidateSourceRef {
	return candidateSourceRef{
		HarnessLoop:       dispatchHarnessLoop,
		SourceWorkflow:    dispatchSourceWorkflow,
		SummaryArtifact:   dispatchSummaryArtifact,
		CandidateArtifact: dispatchCandidateArtifact,
		LiveStatus:        liveStatus,
		SourceGeneratedAt: plan.GeneratedAt,
		SourceExpiresAt:   plan.ExpiresAt,
		Run:               githubRunRecord(),
	}
}
