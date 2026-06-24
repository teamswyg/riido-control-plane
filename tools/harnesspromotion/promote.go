package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func promoteSummary(root string, m manifest, path, out string) error {
	var summary liveSummary
	data, err := os.ReadFile(repoPath(root, path))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &summary); err != nil {
		return err
	}
	source, ok := sourceByID(m, summary.ID)
	if !ok {
		return fmt.Errorf("summary %s is not a promotion source", summary.ID)
	}
	if err := verifySummaryFresh(summary, promotionNow()); err != nil {
		return err
	}
	return writeJSON(repoPath(root, out), buildCandidateEvidence(source, summary))
}

func sourceByID(m manifest, id string) (promotionSource, bool) {
	for _, source := range m.Sources {
		if source.ID == id {
			return source, true
		}
	}
	return promotionSource{}, false
}
