package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/cloudwatchemf/pathutil"
)

func verifySourceChecks(root string, checks []sourceCheck) error {
	for _, check := range checks {
		body, err := os.ReadFile(pathutil.Resolve(root, check.File))
		if err != nil {
			return fmt.Errorf("read source check %s: %w", check.Name, err)
		}
		text := string(body)
		for _, needle := range check.Contains {
			if !strings.Contains(text, needle) {
				return fmt.Errorf("source check %s missing %q in %s", check.Name, needle, check.File)
			}
		}
	}
	return nil
}
