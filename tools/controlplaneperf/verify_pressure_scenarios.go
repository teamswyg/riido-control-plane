package main

import (
	"fmt"
	"strings"
)

func verifyLocalPressureScenarios(root string, scenarios []string) error {
	if len(scenarios) == 0 {
		return fmt.Errorf("performance manifest must declare local pressure scenarios")
	}
	text, err := readText(repoPath(root, "tools/controlplanepressure/scenario.go"))
	if err != nil {
		return err
	}
	for _, scenario := range scenarios {
		if strings.TrimSpace(scenario) == "" {
			return fmt.Errorf("local pressure scenario must not be empty")
		}
		if !strings.Contains(text, `"`+scenario+`"`) {
			return fmt.Errorf("missing local pressure scenario %s", scenario)
		}
	}
	return nil
}
