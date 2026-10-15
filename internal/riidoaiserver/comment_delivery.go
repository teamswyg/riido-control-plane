package riidoaiserver

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const CommentDeliveryKeyHeader = "X-Riido-Comment-Delivery-Key"

type CommentDeliveryConfig struct {
	EvidenceBaseURL string
	Key             string
	HTTPClient      *http.Client
}

type verifiedCommentDelivery struct {
	EventID     string             `json:"event_id"`
	CommentID   string             `json:"comment_id"`
	TaskID      string             `json:"task_id"`
	AgentID     string             `json:"agent_id"`
	WorkspaceID string             `json:"workspace_id"`
	PrincipalID string             `json:"principal_id"`
	Body        string             `json:"body"`
	Context     AIAgentTaskContext `json:"context"`
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

func (s Server) handleCommentDelivery(w http.ResponseWriter, r *http.Request) {
	config := s.config.CommentDelivery
	if config.Key == "" || config.Validate() != nil || s.assignment == nil || s.aiAgent == nil {
		writeError(w, 503, "comment delivery unavailable")
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.Header.Get(CommentDeliveryKeyHeader)), []byte(config.Key)) != 1 {
		writeError(w, 401, "unauthorized")
		return
	}
	var input struct {
		EventID string `json:"event_id"`
	}
	if err := decodeJSON(r, &input); err != nil || strings.TrimSpace(input.EventID) == "" {
		writeError(w, 400, "event_id is required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(config.EvidenceBaseURL, "/")+"/internal/agent-comment-deliveries/"+url.PathEscape(input.EventID), nil)
	if err != nil {
		writeError(w, 502, "evidence request failed")
		return
	}
	req.Header.Set(CommentDeliveryKeyHeader, config.Key)
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	result, err := client.Do(req)
	if err != nil {
		writeError(w, 502, "evidence unavailable")
		return
	}
	defer result.Body.Close()
	if result.StatusCode != http.StatusOK {
		status := http.StatusBadGateway
		if result.StatusCode == http.StatusForbidden || result.StatusCode == http.StatusNotFound || result.StatusCode == http.StatusGone {
			status = http.StatusForbidden
		}
		writeError(w, status, "comment delivery evidence rejected")
		return
	}
	var evidence verifiedCommentDelivery
	if json.NewDecoder(io.LimitReader(result.Body, 2<<20)).Decode(&evidence) != nil || evidence.EventID != input.EventID || evidence.PrincipalID == "" || evidence.WorkspaceID == "" || evidence.TaskID == "" || evidence.AgentID == "" || evidence.CommentID == "" || evidence.Context.Component.ID != evidence.TaskID || (evidence.Context.Component.ComponentType != "task" && evidence.Context.Component.ComponentType != "subtask") {
		writeError(w, 502, "invalid comment delivery evidence")
		return
	}
	principal := AuthorizationResult{PrincipalID: evidence.PrincipalID, WorkspaceID: evidence.WorkspaceID}
	request := AssignAIAgentTaskRequest{AgentID: evidence.AgentID, RequestID: "comment:" + evidence.EventID, Instruction: evidence.Body, SourceCommentID: evidence.CommentID, VerifiedTaskContext: &evidence.Context}
	response, found, lookupErr := s.findTaskRequestReceipt(ctx, principal, evidence.TaskID, request)
	if lookupErr != nil {
		writeCommentDeliveryError(w, lookupErr)
		return
	}
	if found {
		writeJSON(w, http.StatusAccepted, response)
		return
	}
	composed, err := s.assignRequestFromAIAgentClientTaskResult(ctx, principal, "", evidence.TaskID, request)
	if err != nil {
		writeCommentDeliveryError(w, err)
		return
	}
	if composed.IntentGateRequired {
		request.intentGateRequired = true
		response, err = s.aiAgent.CreateAIAgentTaskAgentAssignment(ctx, principal, evidence.TaskID, request)
		if err != nil {
			writeCommentDeliveryError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, response)
		return
	}
	creationContext, creationReceipt := withTaskAgentCreationReceipt(ctx)
	assignment, err := s.assignment.AssignTaskAdditive(creationContext, evidence.TaskID, composed.Request)
	if err != nil {
		writeCommentDeliveryError(w, err)
		return
	}
	request.AssignmentID, request.durableState = assignment.ID, assignmentClientResponseDurableState(assignment)
	if owner, ok := s.assignment.(interface{ taskAgentGenerationSupported() bool }); ok && owner.taskAgentGenerationSupported() && assignmentSequence(assignment.ID) == 0 {
		request.membershipPublicationGeneration = taskAgentAssignmentGeneration(assignment)
	}
	response, err = s.aiAgent.CreateAIAgentTaskAgentAssignment(ctx, principal, evidence.TaskID, request)
	if err != nil {
		s.recoverTaskAgentPublication(ctx, assignment, creationReceipt)
		writeCommentDeliveryError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

// Retry infrastructure and projection failures with the same durable event identity.
func writeCommentDeliveryError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrAIAgentNotFound) {
		writeError(w, http.StatusNotFound, "agent unavailable")
		return
	}
	writeError(w, http.StatusServiceUnavailable, "comment delivery temporarily unavailable")
}
