package main

import "fmt"

type evidenceGraph struct {
	Chains []evidenceChain `json:"chains"`
}

type evidenceChain struct {
	ID       string          `json:"id"`
	NextLoop string          `json:"next_loop"`
	Evidence []evidenceEntry `json:"evidence"`
	Decision string          `json:"decision"`
}

type evidenceEntry struct {
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	Redacted bool   `json:"redacted"`
}

func verifyEvidenceGraph(root string, source intakeSource) error {
	var graph evidenceGraph
	if err := readJSON(repoPath(root, source.EvidenceGraphManifest), &graph); err != nil {
		return err
	}
	for _, chain := range graph.Chains {
		if chain.NextLoop == source.PromotionTarget && chainHasCandidateArtifact(chain, source.CandidateArtifact) {
			return nil
		}
	}
	return fmt.Errorf("candidate artifact %s missing from evidence graph target %s",
		source.CandidateArtifact, source.PromotionTarget)
}

func chainHasCandidateArtifact(chain evidenceChain, artifact string) bool {
	for _, entry := range chain.Evidence {
		if entry.Kind == "artifact" && entry.Path == artifact && entry.Redacted && chain.Decision != "" {
			return true
		}
	}
	return false
}
