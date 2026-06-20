package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var riidoWorkBranchPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*-[0-9]+-.+`)

func validateTargetBranch(branch string) error {
	trimmed := strings.TrimSpace(branch)
	if trimmed != branch {
		return errors.New("target-branch must not have leading or trailing whitespace")
	}
	if strings.Contains(trimmed, "/") {
		return fmt.Errorf("target-branch %q must be a Riido work branchName without path separators", trimmed)
	}
	if !riidoWorkBranchPattern.MatchString(trimmed) {
		return fmt.Errorf("target-branch %q must be the Riido work branchName, for example A-60-AI-Agent-generated-client-handoff", trimmed)
	}
	return nil
}
