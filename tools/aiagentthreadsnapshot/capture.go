package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const schemaVersion = "riido-ai-thread-same-moment-snapshot.v1"

func capture(ctx context.Context, cfg config) (report, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	rep := newReport(cfg, now)
	token := tokenFromEnv(cfg.TokenEnv)
	if token == "" {
		rep.Decision = decisionSummary{
			Status: "blocked_missing_bearer_token",
			Reason: "bearer token environment variable is empty",
		}
		return rep, nil
	}
	client := &http.Client{Timeout: cfg.SSEWindow + 10*time.Second}
	base := "/v2/client/workspaces/" + cfg.WorkspaceID + "/ai-agent"
	v3Path := "/v3/client/workspaces/" + cfg.WorkspaceID + "/ai-agent/tasks/" + cfg.TaskID + "/threads"
	v2Path := base + "/tasks/" + cfg.TaskID + "/threads"
	subPath := base + "/tasks/" + cfg.TaskID + "/thread-stream-subscription"
	v3 := fetchThreads(ctx, client, token, cfg.BaseURL, v3Path, "v3_threads")
	v2 := fetchThreads(ctx, client, token, cfg.BaseURL, v2Path, "v2_threads")
	sub := fetchSubscription(ctx, client, token, cfg.BaseURL, subPath)
	rep.Endpoints = append(rep.Endpoints, v3.Observation, v2.Observation, sub.Observation)
	rep.V3 = summarizeThreads(v3.Payload, cfg.ConversationID)
	rep.V2 = summarizeThreads(v2.Payload, cfg.ConversationID)
	rep.Conversations, rep.ConversationCount = candidateConversations(
		v3.Payload, v2.Payload, 25)
	rep.Subscription = summarizeSubscription(
		sub.Payload, highlightedThreads(rep), cfg.ConversationID)
	if sub.Payload.Stream.Href != "" {
		rep.SSEEvents = captureSSE(ctx, client, token, cfg, sub.Payload.Stream.Href)
	}
	rep.Decision = decide(rep)
	return rep, nil
}

func newReport(cfg config, now string) report {
	return report{
		SchemaVersion: schemaVersion,
		CapturedAt:    now,
		Redacted:      true,
		Source: sourceSummary{
			BaseURL: cfg.BaseURL, WorkspaceID: cfg.WorkspaceID,
			TaskID: cfg.TaskID, ConversationID: cfg.ConversationID,
			TokenEnv: cfg.TokenEnv,
		},
	}
}

func highlightedThreads(rep report) []threadSurface {
	threads := append([]threadSurface{}, rep.V3.HighlightedThreads...)
	return append(threads, rep.V2.HighlightedThreads...)
}

func writeReport(path string, rep report) error {
	body, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}
