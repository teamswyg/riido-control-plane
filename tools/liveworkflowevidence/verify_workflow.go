package main

import (
	"fmt"
	"strings"
)

func verifyWorkflowFile(root string, spec workflowSpec) (int, error) {
	text, err := readText(repoPath(root, spec.Path))
	if err != nil {
		return 0, err
	}
	phrases := []string{
		"go run ./tools/liveworkflowevidence",
		"-workflow " + spec.ID,
		"-evidence-out " + spec.SummaryPath,
		"name: " + spec.SummaryArtifact,
		"path: " + spec.SummaryPath,
		"actions/upload-artifact@v4",
	}
	phrases = append(phrases, spec.SensitiveInputs...)
	phrases = append(phrases, spec.RequiredPhrases...)
	phrases = append(phrases, claimSourcePhrases(spec)...)
	for _, phrase := range phrases {
		if !strings.Contains(text, phrase) {
			return 0, fmt.Errorf("%s missing phrase %q", spec.Path, phrase)
		}
	}
	if strings.Contains(text, "cat out/") || strings.Contains(text, "tee out/") {
		return 0, fmt.Errorf("%s must not write raw shell output into public evidence", spec.Path)
	}
	return len(phrases), nil
}
