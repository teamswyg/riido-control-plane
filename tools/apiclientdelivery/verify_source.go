package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/pathutil"
)

func verifySourceChecks(repoRoot string, checks []sourceCheck, result *verifyResult) error {
	for _, check := range checks {
		if err := verifySourceCheck(repoRoot, check, result); err != nil {
			return err
		}
	}
	return nil
}

func verifySourceCheck(repoRoot string, check sourceCheck, result *verifyResult) error {
	body, err := os.ReadFile(pathutil.Resolve(repoRoot, check.File))
	if err != nil {
		return fmt.Errorf("%s source %s: %w", check.Name, check.File, err)
	}
	text := string(body)
	for _, phrase := range check.Contains {
		result.PhraseChecks++
		if !strings.Contains(text, phrase) {
			return fmt.Errorf("%s source %s missing %q", check.Name, check.File, phrase)
		}
	}
	return nil
}
