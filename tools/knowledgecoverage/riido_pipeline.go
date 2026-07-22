package main

import (
	"encoding/json"
	"strings"
)

type riidoPipeline struct {
	SchemaVersion string `json:"schema_version"`
	Status        string `json:"status"`
	Visibility    string `json:"visibility"`
	Execution     struct {
		Attestation string `json:"attestation"`
	} `json:"execution"`
	Evidence struct {
		Artifact string `json:"artifact"`
	} `json:"evidence_contract"`
	Steps []struct {
		Kind      string   `json:"kind"`
		Command   string   `json:"command"`
		Paths     []string `json:"paths"`
		Redaction string   `json:"redaction"`
		RunWhen   string   `json:"run_when"`
	} `json:"steps"`
}

func readRiidoPipeline(root, path string) (riidoPipeline, bool) {
	if !strings.HasSuffix(path, ".riido.json") {
		return riidoPipeline{}, false
	}
	data, err := readWorkflow(root, path)
	if err != nil {
		return riidoPipeline{}, false
	}
	var pipeline riidoPipeline
	if json.Unmarshal(data, &pipeline) != nil || pipeline.SchemaVersion != "riido-ci-pipeline.v1" ||
		pipeline.Status != "active" || pipeline.Visibility != "private" ||
		pipeline.Execution.Attestation != "required" || pipeline.Evidence.Artifact == "" {
		return riidoPipeline{}, false
	}
	return pipeline, true
}

func riidoPipelineCommands(pipeline riidoPipeline) []string {
	var commands []string
	for _, step := range pipeline.Steps {
		if step.Kind == "shell" && step.Command != "" {
			commands = append(commands, step.Command)
		}
	}
	return commands
}

func riidoPipelineUploadsStrict(pipeline riidoPipeline, artifact, path string) bool {
	if pipeline.Evidence.Artifact != artifact {
		return false
	}
	for _, step := range pipeline.Steps {
		if step.Kind == "artifact" && step.Redaction == "required" && step.RunWhen == "always" {
			for _, candidate := range step.Paths {
				if candidate == path {
					return true
				}
			}
		}
	}
	return false
}
