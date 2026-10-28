package riidoaiserver

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

const CommentDeliveryKeyHeader = "X-Riido-Comment-Delivery-Key"

type CommentDeliveryConfig struct {
	EvidenceBaseURL string
	Key             string
	HTTPClient      *http.Client
}

type verifiedCommentDelivery struct {
	CommentCreatedAt       time.Time                       `json:"comment_created_at"`
	EventID                string                          `json:"event_id"`
	CommentID              string                          `json:"comment_id"`
	ConversationCommentIDs []string                        `json:"conversation_comment_ids,omitempty"`
	RootCommentID          string                          `json:"root_comment_id"`
	TaskID                 string                          `json:"task_id"`
	AgentID                string                          `json:"agent_id"`
	WorkspaceID            string                          `json:"workspace_id"`
	PrincipalID            string                          `json:"principal_id"`
	Body                   string                          `json:"body"`
	Context                AIAgentTaskContext              `json:"context"`
	ConversationMessages   []verifiedCommentContextMessage `json:"conversation_messages,omitempty"`
}

func (config CommentDeliveryConfig) Validate() error {
	if config.EvidenceBaseURL == "" && config.Key == "" {
		return nil
	}
	endpoint, err := url.Parse(config.EvidenceBaseURL)
	if err != nil {
		return errors.New("invalid comment delivery evidence URL")
	}
	invalidScheme := endpoint.Scheme != "https" && (endpoint.Scheme != "http" || (endpoint.Hostname() != "127.0.0.1" && endpoint.Hostname() != "localhost"))
	if endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" || invalidScheme {
		return errors.New("comment delivery evidence URL must use HTTPS (or loopback HTTP)")
	}
	if len(config.Key) < 32 {
		return errors.New("comment delivery key must contain at least 32 bytes")
	}
	return nil
}
