package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/agentcatalogrbac/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/agentcatalogrbac/setutil"
)

func verifySources(root string, m manifest) error {
	if len(m.SourceChecks) == 0 {
		return fmt.Errorf("source_checks are required")
	}
	for _, check := range m.SourceChecks {
		if err := verifySource(root, check); err != nil {
			return err
		}
	}
	return nil
}

func verifySource(root string, check sourceCheck) error {
	if check.Name == "" || check.File == "" || len(check.Contains) == 0 {
		return fmt.Errorf("invalid source check %+v", check)
	}
	data, err := os.ReadFile(pathutil.Resolve(root, check.File))
	if err != nil {
		return fmt.Errorf("read source check %s: %w", check.Name, err)
	}
	text := string(data)
	for _, token := range check.Contains {
		if !setutil.ContainsText(text, token) {
			return fmt.Errorf("source check %s missing %q", check.Name, token)
		}
	}
	return nil
}
