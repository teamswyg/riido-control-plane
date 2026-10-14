package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type producerManifest struct {
	Sources []producerSource `json:"sources"`
}

type producerSource struct {
	ID                    string   `json:"id"`
	SourceWorkflow        string   `json:"source_workflow"`
	CandidateArtifact     string   `json:"candidate_artifact"`
	HarnessLoop           string   `json:"harness_loop"`
	PromotionTarget       string   `json:"promotion_target"`
	RequiredNextArtifacts []string `json:"required_next_artifacts"`
}

func verifyProducer(root string, source intakeSource) error {
	var producer producerManifest
	if err := readJSON(repoPath(root, source.ProducerManifest), &producer); err != nil {
		return err
	}
	for _, candidate := range producer.Sources {
		if candidate.ID == source.ID {
			return verifyProducerSource(source, candidate)
		}
	}
	return fmt.Errorf("source %s missing in producer manifest", source.ID)
}

func verifyProducerSource(source intakeSource, producer producerSource) error {
	if producer.CandidateArtifact != source.CandidateArtifact ||
		producer.SourceWorkflow != source.SourceWorkflow ||
		producer.HarnessLoop != source.HarnessLoop ||
		producer.PromotionTarget != source.PromotionTarget {
		return fmt.Errorf("producer source %s does not match intake source", source.ID)
	}
	return verifyRequiredNextArtifacts(producer.RequiredNextArtifacts, source.ID)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
