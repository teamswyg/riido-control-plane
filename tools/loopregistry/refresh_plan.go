package main

import (
	"path/filepath"
	"time"
)

func refreshPlans(loops []loopRecord, cadence map[string]int) []refreshPlan {
	return refreshPlansAt(loops, cadence, time.Time{})
}

func refreshPlansAt(loops []loopRecord, cadence map[string]int, generatedAt time.Time) []refreshPlan {
	plans := []refreshPlan{}
	for _, loop := range loops {
		minutes := cadence[loop.ID]
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
			EvidenceArtifacts:    redactedEvidenceArtifacts(loop.Evidence),
			EvidenceRefreshes:    evidenceRefreshes(loop),
		})
	}
	return plans
}

func refreshCommand(workflow string) string {
	return "gh workflow run " + filepath.Base(workflow) + " --ref main"
}

func redactedEvidenceArtifacts(values []evidenceSource) []string {
	out := []string{}
	for _, value := range values {
		if value.Redacted {
			out = append(out, value.Path)
		}
	}
	return out
}

func evidenceRefreshes(loop loopRecord) []evidenceArtifactRefresh {
	out := []evidenceArtifactRefresh{}
	for _, source := range loop.Evidence {
		if !source.Redacted {
			continue
		}
		workflow := evidenceRefreshWorkflow(loop, source)
		out = append(out, evidenceArtifactRefresh{
			Artifact:             source.Path,
			RefreshWorkflow:      workflow,
			WorkflowFile:         filepath.Base(workflow),
			ManualRefreshCommand: refreshCommand(workflow),
		})
	}
	return out
}

func evidenceRefreshWorkflow(loop loopRecord, source evidenceSource) string {
	if source.RefreshWorkflow != "" {
		return source.RefreshWorkflow
	}
	return loop.RefreshWorkflow
}
