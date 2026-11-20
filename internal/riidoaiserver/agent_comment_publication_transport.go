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

	"github.com/teamswyg/riido-contracts/assignment"
)

func (s Server) publishAgentComment(ctx context.Context, event TaskEvent) error {
	cfg := s.config.CommentDelivery
	if !cfg.PublishComments || event.Type != EventAssignmentCompleted || event.State != AssignmentCompleted ||
		assignmentMetadataNeedsInput(event.Metadata) || assignment.RecoveryContextFromMetadata(event.Metadata) != nil ||
		event.CreatedBy == "" || event.WorkspaceID == "" {
		return nil
	}
	source, ok := s.aiAgent.(agentCommentPublicationReader)
	if !ok {
		return fmt.Errorf("Agent comment source unavailable")
	}
	publication, err := source.agentCommentPublication(ctx, event)
	if err != nil || publication.MessageID == "" {
		return err
	}
	payload, err := json.Marshal(publication)
	if err != nil {
		return err
	}
	delivery, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(delivery, http.MethodPost,
		strings.TrimRight(cfg.EvidenceBaseURL, "/")+"/internal/agent-comment-publications", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set(CommentDeliveryKeyHeader, cfg.Key)
	req.Header.Set("Content-Type", "application/json")
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	switch res.StatusCode {
	case http.StatusAccepted, http.StatusNotFound:
		// A removed task, inaccessible requester, or deleted source has no audience.
		return nil
	default:
		return fmt.Errorf("Agent comment publication returned HTTP %d", res.StatusCode)
	}
}
