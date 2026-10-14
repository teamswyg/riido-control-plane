package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func promoteSummary(root string, m manifest, path, out string) (candidateEvidence, error) {
	var summary liveSummary
	data, err := os.ReadFile(repoPath(root, path))
	if err != nil {
		return candidateEvidence{}, err
	}
	if err := json.Unmarshal(data, &summary); err != nil {
		return candidateEvidence{}, err
	}
	source, ok := sourceByID(m, summary.ID)
	if !ok {
		return candidateEvidence{}, fmt.Errorf("summary %s is not a promotion source", summary.ID)
	}
	if err := verifySummaryFresh(summary, promotionNow()); err != nil {
		return candidateEvidence{}, err
	}
	candidates := buildCandidateEvidence(source, summary)
	if err := writeJSON(repoPath(root, out), candidates); err != nil {
		return candidateEvidence{}, err
	}
	return candidates, nil
}

func sourceByID(m manifest, id string) (promotionSource, bool) {
	for _, source := range m.Sources {
		if source.ID == id {
			return source, true
		}
	}
	return promotionSource{}, false
}
