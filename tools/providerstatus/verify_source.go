package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/providerstatus/pathutil"
)

func verifySources(root string, checks []sourceCheck) error {
	if len(checks) == 0 {
		return fmt.Errorf("source checks are required")
	}
	for _, check := range checks {
		body, err := os.ReadFile(pathutil.Resolve(root, check.File))
		if err != nil {
			return fmt.Errorf("read source check %q: %w", check.Name, err)
		}
		text := string(body)
		for _, needle := range check.Contains {
			if !strings.Contains(text, needle) {
				return fmt.Errorf("source check %q missing %q in %s", check.Name, needle, check.File)
			}
		}
	}
	return nil
}
