package main

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func captureSSE(ctx context.Context, client *http.Client, token string, cfg config, href string) []sseEventSummary {
	ctx, cancel := context.WithTimeout(ctx, cfg.SSEWindow)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, joinURL(cfg.BaseURL, href), nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	return readSSE(ctx, resp)
}

func readSSE(ctx context.Context, resp *http.Response) []sseEventSummary {
	var events []sseEventSummary
	scanner := bufio.NewScanner(resp.Body)
	var name string
	var data strings.Builder
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return events
		default:
		}
		line := scanner.Text()
		if line == "" {
			events = appendSSE(events, name, data.String())
			name = ""
			data.Reset()
			continue
		}
		if strings.HasPrefix(line, "event:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		}
		if strings.HasPrefix(line, "data:") {
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	time.Sleep(0)
	return events
}

func appendSSE(events []sseEventSummary, name, raw string) []sseEventSummary {
	if raw == "" {
		return events
	}
	var payload sseProgressPayload
	_ = json.Unmarshal([]byte(raw), &payload)
	return append(events, sseEventSummary{
		Event: name, TaskID: payload.TaskID, ThreadID: payload.ThreadID,
		ConversationID: payload.ConversationID, AssignmentID: payload.AssignmentID,
		RunID: payload.RunID, WorkStatus: payload.WorkStatus,
		AssignmentState: payload.AssignmentState, LineCount: len(payload.Lines),
	})
}
