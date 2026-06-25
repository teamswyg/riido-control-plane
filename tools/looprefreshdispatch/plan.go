package main

import (
	"sort"
	"strings"
)

func buildDispatchPlan(root string, source refreshCommandEvidence) (dispatchPlan, error) {
	out := dispatchPlan{
		SchemaVersion: dispatchPlanSchema,
		Status:        "no_dispatch_required",
		SourceStatus:  strings.TrimSpace(source.Status),
	}
	byWorkflow := map[string]map[string]int{}
	ignoredKinds := map[string]bool{}
	for _, command := range source.Commands {
		if command.Kind != "refresh_workflow" {
			out.IgnoredCommandCount++
			ignoredKinds[command.Kind] = true
			continue
		}
		workflow, err := parseRefreshWorkflowCommand(root, command.Command)
		if err != nil {
			return dispatchPlan{}, err
		}
		if byWorkflow[workflow] == nil {
			byWorkflow[workflow] = map[string]int{}
		}
		byWorkflow[workflow][command.LoopID]++
	}
	out.IgnoredCommandKinds = sortedKeys(ignoredKinds)
	out.Dispatches = dispatchesFromWorkflowMap(byWorkflow)
	out.DispatchCount = len(out.Dispatches)
	if out.DispatchCount > 0 {
		out.Status = "dispatch_required"
	}
	return out, nil
}

func dispatchesFromWorkflowMap(byWorkflow map[string]map[string]int) []workflowDispatch {
	workflows := make([]string, 0, len(byWorkflow))
	for workflow := range byWorkflow {
		workflows = append(workflows, workflow)
	}
	sort.Strings(workflows)
	out := make([]workflowDispatch, 0, len(workflows))
	for _, workflow := range workflows {
		loopIDs := sortedKeysInt(byWorkflow[workflow])
		out = append(out, workflowDispatch{
			WorkflowFile: workflow,
			LoopIDs:      loopIDs,
			CommandCount: commandCount(byWorkflow[workflow]),
		})
	}
	return out
}
