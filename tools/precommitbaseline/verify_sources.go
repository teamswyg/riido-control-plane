package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyPreCommitConfig(root string, m manifest, result *verifyResult) error {
	text, err := readText(repoPath(root, m.PreCommitConfig))
	if err != nil {
		return fmt.Errorf("read pre-commit config: %w", err)
	}
	for _, hook := range m.Hooks {
		if hook.ID == "" || hook.Summary == "" || len(hook.Contains) == 0 {
			return fmt.Errorf("hook identity, summary, and contains are required")
		}
		if err := requirePhrases(text, hook.Contains, "hook "+hook.ID, result); err != nil {
			return err
		}
	}
	return nil
}

func verifyScripts(root string, m manifest, result *verifyResult) error {
	for _, script := range m.Scripts {
		if script.Path == "" || script.Summary == "" || len(script.Contains) == 0 {
			return fmt.Errorf("script path, summary, and contains are required")
		}
		text, err := readText(repoPath(root, script.Path))
		if err != nil {
			return fmt.Errorf("read script %s: %w", script.Path, err)
		}
		if err := requirePhrases(text, script.Contains, "script "+script.Path, result); err != nil {
			return err
		}
	}
	return nil
}

func readText(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

func requirePhrases(text string, phrases []string, label string, result *verifyResult) error {
	for _, phrase := range phrases {
		result.PhraseChecks++
		if !strings.Contains(text, phrase) {
			return fmt.Errorf("%s missing phrase %q", label, phrase)
		}
	}
	return nil
}
