package main

import (
	"fmt"
	"strings"
)

func verifyPprofCommands(commands []string) error {
	joined := strings.Join(commands, "\n")
	for _, needle := range []string{"127.0.0.1:6060", "/profile", "/heap", "/goroutine"} {
		if !strings.Contains(joined, needle) {
			return fmt.Errorf("pprof commands must include %q", needle)
		}
	}
	if strings.Contains(joined, "0.0.0.0") {
		return fmt.Errorf("pprof commands must stay loopback-only")
	}
	return nil
}

func verifyLoop(loop loopSpec) error {
	fields := []string{loop.Observation, loop.Hypothesis, loop.Execute, loop.Evaluate, loop.Retrospective}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("audit loop must be complete")
		}
	}
	return nil
}
