package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s Server) fetchCommentDeliveryEvidence(ctx context.Context, w http.ResponseWriter, eventID string) (verifiedCommentDelivery, bool) {
	evidence, err := s.readCommentDeliveryEvidence(ctx, eventID)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, ErrAuthorizationForbidden) {
			status = http.StatusForbidden
		}
		writeError(w, status, "comment delivery evidence unavailable")
		return verifiedCommentDelivery{}, false
	}
	return evidence, true
}

func (s Server) readCommentDeliveryEvidence(ctx context.Context, eventID string) (verifiedCommentDelivery, error) {
	config := s.config.CommentDelivery
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(config.EvidenceBaseURL, "/")+"/internal/agent-comment-deliveries/"+url.PathEscape(eventID), nil)
	if err != nil {
		return verifiedCommentDelivery{}, err
	}
	req.Header.Set(CommentDeliveryKeyHeader, config.Key)
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	result, err := client.Do(req)
	if err != nil {
		return verifiedCommentDelivery{}, err
	}
	defer result.Body.Close()
	if result.StatusCode != http.StatusOK {
		if result.StatusCode == http.StatusForbidden || result.StatusCode == http.StatusNotFound || result.StatusCode == http.StatusGone {
			return verifiedCommentDelivery{}, ErrAuthorizationForbidden
		}
		return verifiedCommentDelivery{}, errors.New("comment evidence temporarily unavailable")
	}

	var evidence verifiedCommentDelivery
	if json.NewDecoder(io.LimitReader(result.Body, 2<<20)).Decode(&evidence) != nil || evidence.EventID != eventID || evidence.PrincipalID == "" || evidence.WorkspaceID == "" || evidence.TaskID == "" || evidence.AgentID == "" || evidence.CommentID == "" || evidence.Context.Component.ID != evidence.TaskID || (evidence.Context.Component.ComponentType != "task" && evidence.Context.Component.ComponentType != "subtask") {
		return verifiedCommentDelivery{}, errors.New("invalid comment delivery evidence")
	}
	return evidence, nil
}
