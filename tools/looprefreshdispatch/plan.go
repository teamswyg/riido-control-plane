package main

import "strings"

func buildDispatchPlan(root string, source refreshCommandEvidence) (dispatchPlan, error) {
	if source.Status == "source_stale" {
		return staleSourceDispatchPlan(source), nil
	}
	if err := verifySourceFresh(source, evidenceNow()); err != nil {
		return dispatchPlan{}, err
	}
	if err := verifySourceCommands(source); err != nil {
		return dispatchPlan{}, err
	}
	generatedAt, expiresAt := evidenceWindow()
	out := dispatchPlan{
		SchemaVersion:      dispatchPlanSchema,
		Status:             "no_dispatch_required",
		GeneratedAt:        generatedAt,
		ExpiresAt:          expiresAt,
		SourceStatus:       strings.TrimSpace(source.Status),
		SourceGeneratedAt:  strings.TrimSpace(source.GeneratedAt),
		SourceExpiresAt:    strings.TrimSpace(source.ExpiresAt),
		SourceCommandCount: source.CommandCount,
		SourceStaleCount:   len(source.StaleSources),
		SourceStaleSources: source.StaleSources,
	}
	if out.SourceStatus == "fresh" {
		return out, nil
	}
	byWorkflow := map[string]workflowScope{}
	ignoredKinds := map[string]bool{}
	for _, command := range source.Commands {
		if command.Kind != "refresh_workflow" {
			out.IgnoredCommandCount++
			ignoredKinds[command.Kind] = true
			out.IgnoredCommands = append(out.IgnoredCommands, command)
			continue
		}
		workflow, err := parseRefreshWorkflowCommand(root, command.Command)
		if err != nil {
			return dispatchPlan{}, err
		}
		scope, ok := byWorkflow[workflow]
		if !ok {
			scope = newWorkflowScope()
		}
		scope.add(command)
		byWorkflow[workflow] = scope
	}
	out.IgnoredCommandKinds = sortedKeys(ignoredKinds)
	out.Dispatches = dispatchesFromWorkflowMap(byWorkflow)
	out.DispatchCount = len(out.Dispatches)
	if out.DispatchCount > 0 {
		out.Status = "dispatch_required"
	}
	return out, nil
}
