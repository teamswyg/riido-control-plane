package main

import "os"

func newCandidateRun() runRecord {
	return runRecord{
		ID:      getenvDefault("GITHUB_RUN_ID", "loop-closure-audit-residual-gaps"),
		Attempt: getenvDefault("GITHUB_RUN_ATTEMPT", "1"),
		SHA:     getenvDefault("GITHUB_SHA", "local"),
		RefName: getenvDefault("GITHUB_REF_NAME", "local"),
		Event:   getenvDefault("GITHUB_EVENT_NAME", "local"),
	}
}

func getenvDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
