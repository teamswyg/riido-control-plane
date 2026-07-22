package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type riidoPipeline struct {
	SchemaVersion string `json:"schema_version"`
	Status        string `json:"status"`
	Repo          string `json:"repo"`
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

func appendRiidoPipelineSources(root string, sources []workflowSource, paths []string) ([]workflowSource, error) {
	for _, path := range paths {
		source, err := riidoPipelineSource(root, path)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

func riidoPipelineSource(root, path string) (workflowSource, error) {
	data, err := os.ReadFile(repoPath(root, path))
	if err != nil {
		return workflowSource{}, fmt.Errorf("read riido-ci pipeline %s: %w", path, err)
	}
	var pipeline riidoPipeline
	if err := json.Unmarshal(data, &pipeline); err != nil {
		return workflowSource{}, fmt.Errorf("decode riido-ci pipeline %s: %w", path, err)
	}
	if pipeline.SchemaVersion != "riido-ci-pipeline.v1" || pipeline.Status != "active" ||
		pipeline.Repo != "riido-control-plane" || pipeline.Visibility != "private" ||
		pipeline.Execution.Attestation != "required" || pipeline.Evidence.Artifact == "" {
		return workflowSource{}, fmt.Errorf("riido-ci pipeline %s is not an active attested private route", path)
	}
	var commands, uploads []string
	for _, step := range pipeline.Steps {
		if step.Kind == "shell" && step.Command != "" {
			commands = append(commands, "- run: "+step.Command)
		}
		if step.Kind == "artifact" && step.Redaction == "required" && step.RunWhen == "always" {
			uploads = append(uploads, step.Paths...)
		}
	}
	if len(commands) == 0 || len(uploads) == 0 {
		return workflowSource{}, fmt.Errorf("riido-ci pipeline %s lacks strict redacted evidence steps", path)
	}
	return workflowSource{Text: strings.Join(commands, "\n"), UploadPaths: uniqueStrings(uploads)}, nil
}
