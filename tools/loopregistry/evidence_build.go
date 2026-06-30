package main

import "time"

func newEvidence(m manifest, result verifyResult, impact *impactEvidence) evidence {
	now := evidenceNow()
	generatedAt, expiresAt := evidenceWindowAt(now, loopRegistryEvidenceTTLHours)
	index := architectureIndexFor(m.Claims, result.ClaimSurfaces)
	attachTargetVerifierPlan(impact, index, result.ClaimSurfaces)
	refreshPlans := evidenceRefreshPlans(m, result, now)
	return evidence{
		SchemaVersion:             evidenceSchema,
		ID:                        m.ID,
		Status:                    "verified",
		GeneratedAt:               generatedAt,
		ExpiresAt:                 expiresAt,
		LoopCount:                 result.Loops,
		HarnessCount:              result.Harnesses,
		ClosedLoopCount:           result.ClosedLoops,
		ClaimCount:                result.Claims,
		GraphEdgeCount:            result.GraphEdges,
		MaxExpiryHours:            result.MaxExpiryHours,
		SemanticHashes:            result.Hashes,
		EvidenceKinds:             m.EvidenceKinds,
		LoopSurfaces:              loopSurfaces(m.Loops),
		CoverageDimensions:        loopCoverageDimensionSurfaces(),
		ArchitectureIndex:         index,
		EvidenceGraph:             m.EvidenceGraph,
		ClaimSurfaces:             result.ClaimSurfaces,
		RefreshWorkflows:          refreshWorkflows(m.Loops),
		RefreshCadenceMinutes:     result.RefreshCadenceMinutes,
		RefreshPlanSummary:        summarizeRefreshPlans(refreshPlans),
		RefreshPlans:              refreshPlans,
		ProviderCoverage:          providerCoverage(m.Loops),
		HarnessWorkflowExclusions: m.HarnessWorkflowExclusions,
		HarnessPromotionWorkflows: result.HarnessPromotionWorkflows,
		HarnessCandidateArtifacts: result.HarnessCandidateArtifacts,
		Workflow:                  m.Workflow,
		GeneratedDoc:              m.GeneratedDoc,
		EvidenceArtifact:          m.EvidenceArtifact,
		Loop:                      m.Loop,
		Impact:                    impact,
	}
}

func evidenceRefreshPlans(m manifest, result verifyResult, now time.Time) []refreshPlan {
	return refreshPlansAt(
		m.Loops,
		m.Claims,
		result.ClaimSurfaces,
		result.RefreshCadenceMinutes,
		now,
	)
}
