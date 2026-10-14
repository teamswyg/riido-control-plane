package main

import (
	"path/filepath"
	"time"
)

func refreshPlans(
	loops []loopRecord,
	claims []claimBinding,
	surfaces []claimSurface,
	cadence map[string]int,
) []refreshPlan {
	return refreshPlansAt(loops, claims, surfaces, cadence, time.Time{})
}

func refreshPlansAt(
	loops []loopRecord,
	claims []claimBinding,
	surfaces []claimSurface,
	cadence map[string]int,
	generatedAt time.Time,
) []refreshPlan {
	plans := []refreshPlan{}
	coverage := refreshPlanClaimCoverage(claims, surfaces)
	for _, loop := range loops {
		minutes := cadence[loop.ID]
		claims := coverage[loop.ID]
		plans = append(plans, refreshPlan{
			LoopID:               loop.ID,
			Kind:                 loop.Kind,
			RefreshWorkflow:      loop.RefreshWorkflow,
			WorkflowFile:         filepath.Base(loop.RefreshWorkflow),
			CadenceMinutes:       minutes,
			ExpiresAfterHours:    loop.ExpiresAfterHours,
			EvidenceGeneratedAt:  refreshEvidenceGeneratedAt(generatedAt),
			NextRefreshDueAt:     refreshDueAt(generatedAt, minutes),
			EvidenceExpiresAt:    refreshExpiresAt(generatedAt, loop.ExpiresAfterHours),
			ManualRefreshCommand: refreshCommand(loop.RefreshWorkflow),
			ClaimIDs:             claims.ClaimIDs,
			VerifierCommands:     claims.VerifierCommands,
			NextCommands:         refreshPlanNextCommands(loop, claims),
			EvidenceArtifacts:    redactedEvidenceArtifacts(loop.Evidence),
			EvidenceRefreshes:    evidenceRefreshes(loop),
		})
	}
	return plans
}

func refreshCommand(workflow string) string {
	return "gh workflow run " + filepath.Base(workflow) + " --ref main"
}
