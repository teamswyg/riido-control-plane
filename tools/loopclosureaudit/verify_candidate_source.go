package main

import (
	"fmt"
	"strings"
)

func verifyCandidateSources(sources []candidateSource, m manifest) error {
	if len(sources) != 1 {
		return fmt.Errorf("loop closure audit must expose exactly one residual-gap candidate source")
	}
	source := sources[0]
	fields := []string{
		source.ID, source.SourceWorkflow, source.SummaryArtifact, source.CandidateArtifact,
		source.HarnessLoop, source.PromotionTarget,
	}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("residual-gap candidate source must bind workflow, artifacts, loop, and target")
		}
	}
	if source.SourceWorkflow != m.Workflow || source.SummaryArtifact != m.EvidenceArtifact {
		return fmt.Errorf("residual-gap candidate source must use audit workflow and evidence artifact")
	}
	if source.HarnessLoop != "loop_closure_audit" || source.PromotionTarget != "closed_loop_candidate" {
		return fmt.Errorf("residual-gap candidate source must promote loop closure audit gaps")
	}
	return verifyCandidateArtifacts(source.RequiredNextArtifacts)
}

func verifyCandidateArtifacts(values []string) error {
	required := map[string]bool{"claim_binding": false, "decision_record": false, "evidence_graph_edge": false}
	for _, value := range values {
		if _, ok := required[value]; ok {
			required[value] = true
		}
	}
	for artifact, ok := range required {
		if !ok {
			return fmt.Errorf("residual-gap candidate source missing next artifact %s", artifact)
		}
	}
	return nil
}
