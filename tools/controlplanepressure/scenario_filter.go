package main

import "fmt"

func selectedScenarios(cfg config) ([]scenario, error) {
	all := scenarios()
	if len(cfg.ScenarioNames) == 0 {
		return all, nil
	}
	byName := make(map[string]scenario, len(all))
	for _, sc := range all {
		byName[sc.name] = sc
	}
	selected := make([]scenario, 0, len(cfg.ScenarioNames))
	seen := map[string]bool{}
	for _, name := range cfg.ScenarioNames {
		sc, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("unknown scenario %q", name)
		}
		if !seen[name] {
			selected = append(selected, sc)
			seen[name] = true
		}
	}
	return selected, nil
}
