package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// failureStageError names the step that failed without changing the error's
// meaning: Error and Unwrap return the cause, so status mapping and retries
// that use errors.Is/As behave exactly as before.
type failureStageError struct {
	stage string
	err   error
}

func (e failureStageError) Error() string { return e.err.Error() }
func (e failureStageError) Unwrap() error { return e.err }

func withFailureStage(stage string, err error) error {
	if err == nil {
		return nil
	}
	return failureStageError{stage: stage, err: err}
}

// failureStages returns the outermost and innermost recorded stages.
func failureStages(err error) (string, string) {
	var staged failureStageError
	if !errors.As(err, &staged) {
		return "unknown", "unknown"
	}
	outer, inner := staged.stage, staged.stage
	for rest := staged.err; errors.As(rest, &staged); rest = staged.err {
		inner = staged.stage
	}
	return outer, inner
}

// failureKind is a bounded classification for logs; it never includes
// request bodies, comment text, titles or user identifiers.
func failureKind(ctx context.Context, err error) string {
	var readErr taskContextReadError
	var latest latestTaskDocumentError
	switch {
	case err == nil:
		return "none"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case errors.As(err, &readErr):
		return "context_" + boundedTaskContextOutcome(readErr.outcome)
	case errors.Is(err, errAIAgentSnapshotConflict):
		return "snapshot_conflict"
	case errors.Is(err, errTaskAgentSnapshotChanged), errors.Is(err, ErrAIAgentMembershipChanged):
		return "membership_changed"
	case errors.Is(err, ErrAIAgentTaskThreadConflict):
		return "thread_conflict"
	case errors.Is(err, errNativeSessionUnavailable):
		return "session_unavailable"
	case errors.Is(err, errNativeSessionConfigurationMismatch):
		return "session_configuration_mismatch"
	case errors.Is(err, errNativeSessionRuntimeMismatch):
		return "session_runtime_mismatch"
	case isDynamoDBTransactionContention(err):
		return "store_conflict"
	case errors.Is(err, ErrAuthorizationForbidden):
		return "denied"
	case errors.Is(err, ErrAIAgentNotFound):
		return "not_found"
	case errors.Is(err, ErrAIAgentAssigned):
		return "already_assigned"
	case errors.As(err, &latest):
		return "context_unavailable"
	case ctx != nil && errors.Is(ctx.Err(), context.DeadlineExceeded):
		return "timeout"
	case ctx != nil && errors.Is(ctx.Err(), context.Canceled):
		return "cancelled"
	default:
		return "other"
	}
}

func boundedTaskContextOutcome(outcome string) string {
	switch outcome {
	case "denied", "missing", "invalid", "unavailable", "budget_exhausted", "cancelled":
		return outcome
	default:
		return "unknown"
	}
}

// innermostErrorType helps triage the "other" kind; type names carry no data.
func innermostErrorType(err error) string {
	for err != nil {
		next := errors.Unwrap(err)
		if next == nil {
			return strings.TrimPrefix(fmt.Sprintf("%T", err), "*")
		}
		err = next
	}
	return "none"
}

func logCommentDeliveryFailure(ctx context.Context, path, stage, requestID string, started time.Time, err error) {
	if err == nil {
		return
	}
	outer, inner := failureStages(err)
	if stage == "" {
		stage = outer
	}
	traceID, spanID := traceCorrelationFromContext(ctx)
	log.Printf("event=comment_delivery_failed path=%q stage=%q cause_stage=%q failure_kind=%q error_type=%q request_ref=%q duration_ms=%d trace_id=%q span_id=%q",
		path, stage, inner, failureKind(ctx, err), innermostErrorType(err), providerHealthDeviceRef(requestID), time.Since(started).Milliseconds(), traceID, spanID)
}

func logTaskContextUnavailable(ctx context.Context, err error) {
	if err == nil {
		return
	}
	_, inner := failureStages(err)
	traceID, spanID := traceCorrelationFromContext(ctx)
	log.Printf("event=task_context_unavailable stage=%q failure_kind=%q error_type=%q trace_id=%q span_id=%q",
		inner, failureKind(ctx, err), innermostErrorType(err), traceID, spanID)
}

func commentDeliveryPath(request AssignAIAgentTaskRequest) string {
	if call := request.commentInvocation; call != nil && call.OriginAssignmentID != "" {
		if call.HumanFollowup {
			return "human_followup"
		}
		return "agent_handoff"
	}
	return "comment"
}

func (s Server) deliverCommentInvocation(ctx context.Context, principal AuthorizationResult, evidence verifiedCommentDelivery, request AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	started := time.Now()
	response, err := s.deliverCommentInvocationStages(ctx, principal, evidence, request)
	logCommentDeliveryFailure(ctx, commentDeliveryPath(request), "", request.RequestID, started, err)
	return response, err
}
