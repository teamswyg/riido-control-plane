package main

func summarizeRefreshPlans(plans []refreshPlan) refreshPlanSummary {
	summary := refreshPlanSummary{PlanCount: len(plans)}
	workflows := []string{}
	for _, plan := range plans {
		workflows = appendUnique(workflows, plan.RefreshWorkflow)
		summary.EvidenceArtifactCount += len(plan.EvidenceArtifacts)
		summary.NextCommandCount += len(plan.NextCommands)
		summary.VerifierCommandCount += len(plan.VerifierCommands)
		summary.ClaimBindingCount += len(plan.ClaimIDs)
		if plan.ManualRefreshCommand != "" {
			summary.ManualCommandCount++
		}
	}
	summary.RefreshWorkflowCount = len(workflows)
	return summary
}
