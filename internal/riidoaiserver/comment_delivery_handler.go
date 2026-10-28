package riidoaiserver

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"
)

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
	evidence, ok := s.fetchCommentDeliveryEvidence(ctx, w, input.EventID)
	if !ok {
		return
	}
	principal := AuthorizationResult{PrincipalID: evidence.PrincipalID, WorkspaceID: evidence.WorkspaceID}
	if s.config.AssignmentContext != nil && evidence.Context.Extended == nil {
		evidence.Context.Extended = commentDeliveryExtendedContext(evidence.CommentID, evidence.Body)
	}
	release, err := s.taskAgentExecutionGate.beginPublication(ctx, evidence.TaskID)
	if err != nil {
		writeCommentDeliveryError(w, err)
		return
	}
	defer release()
	request := AssignAIAgentTaskRequest{AgentID: evidence.AgentID, RequestID: "comment:" + evidence.EventID, Instruction: evidence.Body, SourceCommentID: evidence.CommentID, VerifiedTaskContext: &evidence.Context, commentInvocation: newCommentInvocation(evidence)}
	response, err := s.deliverCommentInvocation(ctx, principal, evidence, request)
	if err != nil {
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
