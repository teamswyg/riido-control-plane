package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/pathutil"
)

func verifySourceChecks(repo string, checks []sourceCheck) error {
	if len(checks) == 0 {
		return fmt.Errorf("source checks are required")
	}
	for _, check := range checks {
		if err := verifySourceCheck(repo, check); err != nil {
			return err
		}
	}
	return nil
}

func verifySourceCheck(repo string, check sourceCheck) error {
	if check.Name == "" || check.File == "" || len(check.Contains) == 0 {
		return fmt.Errorf("incomplete source check: %+v", check)
	}
	body, err := os.ReadFile(pathutil.Resolve(repo, check.File))
	if err != nil {
		return fmt.Errorf("read source check %q: %w", check.Name, err)
	}
	for _, needle := range check.Contains {
		if !strings.Contains(string(body), needle) {
			return fmt.Errorf("source check %q missing %q in %s", check.Name, needle, check.File)
		}
	}
	return nil
}
