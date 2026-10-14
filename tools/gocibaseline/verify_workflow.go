package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyWorkflow(root string, m manifest, result *verifyResult) error {
	data, err := os.ReadFile(repoPath(root, m.Workflow))
	if err != nil {
		return fmt.Errorf("read workflow: %w", err)
	}
	text := string(data)
	for _, gate := range m.Gates {
		if gate.ID == "" || gate.Summary == "" || len(gate.Contains) == 0 {
			return fmt.Errorf("gate identity and contains are required")
		}
		for _, phrase := range gate.Contains {
			result.PhraseChecks++
			if !strings.Contains(text, phrase) {
				return fmt.Errorf("gate %q missing phrase %q", gate.ID, phrase)
			}
		}
	}
	return nil
}
