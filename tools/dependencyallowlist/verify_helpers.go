package main

import (
	"sort"
	"strings"
)

func allowedByPath(c contract) map[string]allowedModule {
	allowed := map[string]allowedModule{}
	for _, module := range c.AllowedDirectModules {
		allowed[module.Path] = module
	}
	return allowed
}

func unusedAllowedModules(c contract, used map[string]struct{}) []string {
	var unused []string
	for _, module := range c.AllowedDirectModules {
		if _, ok := used[module.Path]; !ok {
			unused = append(unused, module.Path)
		}
	}
	sort.Strings(unused)
	return unused
}

func sortModules(modules []goModule) {
	sort.Slice(modules, func(i, j int) bool { return modules[i].Path < modules[j].Path })
}

func formatModules(modules []goModule) string {
	var lines []string
	for _, module := range modules {
		line := module.Path
		if module.Version != "" {
			line += " " + module.Version
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
