package main

import (
	"fmt"
	"sort"
	"strings"
)

func verifierCommandsForClaim(
	claim claimBinding,
	testPaths []string,
	tests map[string][]string,
) []string {
	claimPackages := claimBoundPackages(claim.Files)
	byPackage := map[string][]string{}
	for _, verifier := range sortedCopy(claim.Verifiers) {
		for _, pkg := range verifierPackages(verifier, claimPackages, tests) {
			byPackage[pkg] = append(byPackage[pkg], verifier)
		}
	}
	commands := make([]string, 0, len(byPackage))
	for _, pkg := range sortedKeys(byPackage) {
		pattern := strings.Join(sortedCopy(byPackage[pkg]), "|")
		commands = append(commands, fmt.Sprintf("go test %s -run '^(%s)$' -count=1", pkg, pattern))
	}
	return commands
}

func sortedKeys(values map[string][]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func verifierPackages(verifier string, claimPackages map[string]bool, tests map[string][]string) []string {
	all := sortedCopy(tests[verifier])
	if len(claimPackages) == 0 {
		return all
	}
	out := []string{}
	for _, pkg := range all {
		if claimPackages[pkg] {
			out = append(out, pkg)
		}
	}
	if len(out) > 0 {
		return out
	}
	return all
}
