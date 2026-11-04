package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func (s Server) pushAgentCommentCount(ctx context.Context, key, workspaceID, taskID string) {
	observedAt := time.Now().UTC()
	// The list count is task-wide, not per viewer, so read every thread of the task.
	principal := AuthorizationResult{WorkspaceID: workspaceID, Roles: []AgentCatalogRole{AgentCatalogRoleAdmin}}
	history, err := s.aiAgent.ListAIAgentTaskThreadHistory(ctx, principal, taskID)
	if err != nil {
		slog.WarnContext(ctx, "agent comment count pending", "task_id", taskID, "error", err)
		return
	}
	count := countVisibleAgentCommentMessages(history.Threads)
	if s.agentCommentCounts.isAlreadySent(key, count) {
		return
	}
	body := map[string]any{"workspace_id": workspaceID, "task_id": taskID, "message_count": count, "observed_at": observedAt}
	if err := s.writeAgentCommentCount(ctx, body); err != nil {
		slog.WarnContext(ctx, "agent comment count pending", "task_id", taskID, "error", err)
		return
	}
	s.agentCommentCounts.markSent(key, count)
}

func (s Server) writeAgentCommentCount(ctx context.Context, body any) error {
	config := s.config.CommentDelivery
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, strings.TrimRight(config.EvidenceBaseURL, "/")+agentCommentCountPath, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set(CommentDeliveryKeyHeader, config.Key)
	req.Header.Set("Content-Type", "application/json")
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("agent comment count returned %d", response.StatusCode)
	}
	return nil
}
