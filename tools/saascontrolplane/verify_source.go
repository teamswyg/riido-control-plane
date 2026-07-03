package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/pathutil"
)

func verifySourceCheck(repo, boundaryID string, check sourceCheck) error {
	if check.Name == "" || check.File == "" || len(check.Contains) == 0 {
		return fmt.Errorf("boundary %q has incomplete source check: %+v", boundaryID, check)
	}
	body, err := os.ReadFile(pathutil.RepoPath(repo, check.File))
	if err != nil {
		return fmt.Errorf("read source check %q/%q: %w", boundaryID, check.Name, err)
	}
	text := string(body)
	for _, needle := range check.Contains {
		if !strings.Contains(text, needle) {
			return fmt.Errorf("source check %q/%q missing %q in %s", boundaryID, check.Name, needle, check.File)
		}
	}
	return nil
}

func verifyRequiredPhrases(repo, generatedDoc, generatedText string, phrases []phrase) error {
	for _, phrase := range phrases {
		text := generatedText
		if phrase.File != generatedDoc {
			body, err := os.ReadFile(pathutil.RepoPath(repo, phrase.File))
			if err != nil {
				return fmt.Errorf("read required phrase file %q: %w", phrase.File, err)
			}
			text = string(body)
		}
		if !strings.Contains(text, phrase.Contains) {
			return fmt.Errorf("required phrase %q missing from %s", phrase.Contains, phrase.File)
		}
	}
	return nil
}
