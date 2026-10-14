package main

import "os"

func githubRunRecord() pressureRunRecord {
	run := pressureRunRecord{
		ID:      getenvDefault("GITHUB_RUN_ID", "local"),
		Attempt: os.Getenv("GITHUB_RUN_ATTEMPT"),
		SHA:     os.Getenv("GITHUB_SHA"),
		RefName: os.Getenv("GITHUB_REF_NAME"),
		Event:   os.Getenv("GITHUB_EVENT_NAME"),
	}
	return run
}

func getenvDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
