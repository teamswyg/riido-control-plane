package main

import (
	"fmt"
	"strings"
)

func verifyCommands(m manifest) error {
	required := map[string]string{
		m.BenchmarkCommand:         "-benchmem",
		m.LocalPressureCommand:     "go run ./tools/controlplanepressure",
		m.ManualPressureCommand:    "-concurrency 1,8,32,128",
		m.LocalPprofCommand:        "-pprof-dir",
		m.ArchitectureQueryCommand: "-architecture-query-out",
		m.RaceCommand:              "go test -race",
		m.PprofCommand:             "127.0.0.1:6060",
		m.LiveLoadCommand:          "go run ./tools/aiagentload",
	}
	for command, needle := range required {
		if !strings.Contains(command, needle) {
			return fmt.Errorf("performance command missing %q: %s", needle, command)
		}
	}
	return verifyPressureCandidateCommands(m)
}

func verifyPressureCandidateCommands(m manifest) error {
	for _, command := range []string{
		m.LocalPressureCommand,
		m.ManualPressureCommand,
		m.LocalPprofCommand,
	} {
		if !strings.Contains(command, "-candidate-out") {
			return fmt.Errorf("performance pressure command missing candidate output: %s", command)
		}
	}
	return nil
}
