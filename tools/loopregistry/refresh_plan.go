package main

import "path/filepath"

func refreshPlans(loops []loopRecord, cadence map[string]int) []refreshPlan {
	plans := []refreshPlan{}
	for _, loop := range loops {
		plans = append(plans, refreshPlan{
			LoopID:               loop.ID,
			Kind:                 loop.Kind,
			RefreshWorkflow:      loop.RefreshWorkflow,
			WorkflowFile:         filepath.Base(loop.RefreshWorkflow),
			CadenceMinutes:       cadence[loop.ID],
			ExpiresAfterHours:    loop.ExpiresAfterHours,
			ManualRefreshCommand: refreshCommand(loop.RefreshWorkflow),
			EvidenceArtifacts:    redactedEvidenceArtifacts(loop.Evidence),
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
