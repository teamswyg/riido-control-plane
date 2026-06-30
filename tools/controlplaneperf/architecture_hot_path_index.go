package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func applyHotPathRows(
	byPath map[string]*architectureFileEvidence,
	paths []hotPath,
) {
	for _, path := range paths {
		for _, file := range path.Files {
			row := architectureFileRow(byPath, file)
			row.HotPathIDs = appendUnique(row.HotPathIDs, path.ID)
			row.HotPathCategories = appendUnique(row.HotPathCategories, path.Category)
			row.Benchmarks = appendAllUnique(row.Benchmarks, path.Benchmarks)
			row.Tests = appendAllUnique(row.Tests, path.Tests)
			row.OptimizationCandidates = appendUnique(row.OptimizationCandidates, path.Candidate)
			row.TargetVerifierCommands = appendAllUnique(
				row.TargetVerifierCommands,
				hotPathTargetCommands(file, path),
			)
		}
	}
}

func hotPathTargetCommands(file string, path hotPath) []string {
	pkg := "./" + filepath.ToSlash(filepath.Dir(file))
	commands := []string{}
	if len(path.Benchmarks) > 0 {
		commands = append(commands, fmt.Sprintf(
			"go test %s -run '^$' -bench '%s' -benchmem -benchtime=100ms -count=1",
			pkg,
			symbolRegex(path.Benchmarks),
		))
	}
	if len(path.Tests) > 0 {
		commands = append(commands, fmt.Sprintf(
			"go test %s -run '%s' -count=1",
			pkg,
			symbolRegex(path.Tests),
		))
	}
	return commands
}

func symbolRegex(symbols []string) string {
	return "^(" + strings.Join(symbols, "|") + ")$"
}
