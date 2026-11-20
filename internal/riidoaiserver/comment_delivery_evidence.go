package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s Server) readCommentDeliveryEvidence(ctx context.Context, eventID string) (verifiedCommentDelivery, error) {
	config := s.config.CommentDelivery
	evidence, err := s.fetchInvocationEvidence(ctx, strings.TrimRight(config.EvidenceBaseURL, "/")+"/internal/agent-comment-deliveries/"+url.PathEscape(eventID))
	if err != nil {
		return verifiedCommentDelivery{}, err
	}
	if evidence.EventID != eventID || evidence.AgentID == "" || evidence.CommentID == "" {
		return verifiedCommentDelivery{}, errors.New("invalid comment delivery evidence")
	}
	return evidence, nil
}

func (s Server) fetchInvocationEvidence(ctx context.Context, endpoint string) (verifiedCommentDelivery, error) {
	config := s.config.CommentDelivery
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return verifiedCommentDelivery{}, err
	}
	req.Header.Set(CommentDeliveryKeyHeader, config.Key)
	schema := s.invocationReadScopeSchema()
	if schema != "" {
		req.Header.Set(commentReadScopeVersionHeader, schema)
	}
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

	evidence, decodeErr := decodeCommentDeliveryEvidence(result.Body)
	if decodeErr != nil || evidence.PrincipalID == "" || evidence.WorkspaceID == "" || evidence.TaskID == "" || evidence.Context.Component.ID != evidence.TaskID || (evidence.Context.Component.ComponentType != "task" && evidence.Context.Component.ComponentType != "subtask") {
		return verifiedCommentDelivery{}, errors.New("invalid comment delivery evidence")
	}
	if schema == taskReadScopeNativeSchema && (evidence.Context.ReadScopeEvidence == nil || evidence.Context.ReadScopeEvidence.Scope == nil || evidence.Context.ReadScopeEvidence.Scope.SchemaVersion != schema) {
		return verifiedCommentDelivery{}, errors.New("native scoped comment evidence required")
	}
	if err := s.prepareCommentReadScope(ctx, &evidence); err != nil {
		return verifiedCommentDelivery{}, errors.New("invalid scoped comment evidence")
	}
	return evidence, nil
}
