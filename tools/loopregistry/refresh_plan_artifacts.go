package main

import "path/filepath"

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
