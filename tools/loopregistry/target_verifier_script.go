package main

import (
	"fmt"
	"os"
	"strings"
)

func writeTargetVerifierScript(path string, impact *impactEvidence) error {
	if impact == nil || impact.TargetVerifierPlan == nil {
		return fmt.Errorf("target verifier script requires impact target verifier plan")
	}
	body, err := targetVerifierScript(impact.TargetVerifierPlan)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o755)
}

func targetVerifierScript(plan *targetVerifierPlan) (string, error) {
	commands := targetVerifierScriptCommands(plan)
	if len(commands) == 0 {
		return "", fmt.Errorf("target verifier script requires at least one command")
	}
	var b strings.Builder
	b.WriteString("#!/usr/bin/env bash\nset -euo pipefail\n\n")
	for _, command := range commands {
		if strings.ContainsAny(command, "\x00\n\r") {
			return "", fmt.Errorf("target verifier command contains newline or NUL")
		}
		b.WriteString(command)
		b.WriteString("\n")
	}
	return b.String(), nil
}

func targetVerifierScriptCommands(plan *targetVerifierPlan) []string {
	if plan == nil {
		return nil
	}
	if len(plan.FocusedCommands) > 0 {
		return append([]string(nil), plan.FocusedCommands...)
	}
	return append([]string(nil), plan.EntrypointCommands...)
}
