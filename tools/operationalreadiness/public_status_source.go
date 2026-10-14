package main

import "os"

type publicStatusSource struct {
	Workflow string
	Commit   string
	RunID    string
	RunURL   string
}

func currentPublicStatusSource() publicStatusSource {
	source := publicStatusSource{
		Workflow: envOrDefault("GITHUB_WORKFLOW", "local"),
		Commit:   envOrDefault("GITHUB_SHA", "local"),
		RunID:    envOrDefault("GITHUB_RUN_ID", "local"),
	}
	server := os.Getenv("GITHUB_SERVER_URL")
	repo := os.Getenv("GITHUB_REPOSITORY")
	if server != "" && repo != "" && source.RunID != "local" {
		source.RunURL = server + "/" + repo + "/actions/runs/" + source.RunID
	}
	return source
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
