package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/agentcatalogrbac/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/agentcatalogrbac/setutil"
)

func evidenceProfileFor(m manifest, id string) (profile, error) {
	for _, candidate := range m.EvidenceProfiles {
		if candidate.ID == id {
			return candidate, nil
		}
	}
	return profile{}, fmt.Errorf("unknown evidence profile %q", id)
}

func verifyProfiles(root string, m manifest) error {
	if len(m.EvidenceProfiles) == 0 {
		return fmt.Errorf("evidence_profiles are required")
	}
	for _, profile := range m.EvidenceProfiles {
		if err := verifyProfile(root, profile); err != nil {
			return err
		}
	}
	return nil
}

func verifyProfile(root string, p profile) error {
	if p.ID == "" || p.Workflow == "" || p.EvidenceArtifact == "" || p.Focus == "" || len(p.TestPatterns) == 0 {
		return fmt.Errorf("invalid evidence profile %+v", p)
	}
	body, err := os.ReadFile(pathutil.Resolve(root, p.Workflow))
	if err != nil {
		return fmt.Errorf("read evidence profile workflow %s: %w", p.ID, err)
	}
	text := string(body)
	if !setutil.ContainsText(text, p.EvidenceArtifact) {
		return fmt.Errorf("profile %s workflow missing artifact %s", p.ID, p.EvidenceArtifact)
	}
	for _, pattern := range p.TestPatterns {
		if !setutil.ContainsText(text, pattern) {
			return fmt.Errorf("profile %s workflow missing test pattern %s", p.ID, pattern)
		}
	}
	return nil
}
