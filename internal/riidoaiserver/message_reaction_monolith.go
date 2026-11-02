package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (s Server) writeMonolithReaction(ctx context.Context, method, path string, body any) error {
	config := s.config.CommentDelivery
	if config.EvidenceBaseURL == "" || config.Key == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(config.EvidenceBaseURL, "/")+path, bytes.NewReader(payload))
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
		return monolithReactionStatusError{status: response.StatusCode}
	}
	return nil
}

// monolithReactionStatusError keeps the status so a permanent rejection is
// not retried.
type monolithReactionStatusError struct{ status int }

func (err monolithReactionStatusError) Error() string {
	return fmt.Sprintf("reaction delivery returned %d", err.status)
}

func (err monolithReactionStatusError) permanent() bool {
	return err.status == http.StatusBadRequest || err.status == http.StatusNotFound
}
