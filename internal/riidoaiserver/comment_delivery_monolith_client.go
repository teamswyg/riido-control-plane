package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RIID-5906 monolith internal APIs, called with the comment delivery key.
const commentDeliveryMonolithTimeout = 5 * time.Second

func (config CommentDeliveryConfig) monolithEnabled() bool {
	return config.Key != "" && config.EvidenceBaseURL != "" && config.Validate() == nil
}

// putCommentStatusReaction replaces the agent status reaction on the comment
// of a delivery. "" removes it.
func (config CommentDeliveryConfig) putCommentStatusReaction(ctx context.Context, eventID, emoji string) error {
	return config.sendMonolith(ctx, http.MethodPut, "/internal/agent-comment-deliveries/"+url.PathEscape(eventID)+"/reaction", map[string]string{"emoji": emoji})
}

func (config CommentDeliveryConfig) postThreadReactionNotification(ctx context.Context, notification agentThreadReactionNotification) error {
	return config.sendMonolith(ctx, http.MethodPost, "/internal/agent-thread-reaction-notifications", notification)
}

// monolithStatusError keeps the response status so callers can tell a
// permanent rejection from a retryable failure.
type monolithStatusError struct{ status int }

func (err monolithStatusError) Error() string {
	return fmt.Sprintf("monolith responded %d", err.status)
}

func (err monolithStatusError) permanent() bool {
	return err.status == http.StatusBadRequest || err.status == http.StatusNotFound
}

func (config CommentDeliveryConfig) sendMonolith(ctx context.Context, method, path string, body any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, commentDeliveryMonolithTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(config.EvidenceBaseURL, "/")+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set(CommentDeliveryKeyHeader, config.Key)
	req.Header.Set("Content-Type", "application/json")
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: commentDeliveryMonolithTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return monolithStatusError{status: response.StatusCode}
	}
	return nil
}
