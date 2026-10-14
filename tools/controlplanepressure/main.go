package main

import (
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "controlplanepressure:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	profiler, err := startProfiler(cfg.PprofDir)
	if err != nil {
		return err
	}
	report, err := run(cfg)
	profiles, profileErr := profiler.Stop()
	if err != nil {
		return err
	}
	if profileErr != nil {
		return profileErr
	}
	report.Profiles = profiles
	if err := writePressureCandidateEvidence(cfg.CandidateOut, report); err != nil {
		return err
	}
	printSummary(report)
	return writeReport(cfg.EvidenceOut, report)
}
