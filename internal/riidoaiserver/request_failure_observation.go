package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func logRequestFailure(r *http.Request, stage string, err error) {
	traceID, spanID := traceCorrelationFromContext(r.Context())
	if stage == "authorization" {
		hints := authorizationTokenHints(r, time.Now())
		log.Printf("event=control_plane_request_failed route=%q stage=%q error_type=%q trace_id=%q span_id=%q request_id=%q token_source=%q token_format=%q expiry_hint=%q token_header_relation=%q", sseLogRoute(r), boundedRequestFailureStage(stage), boundedRequestFailureType(err), traceID, spanID, requestCorrelationID(r), hints.source, hints.format, hints.expiryHint, hints.headerRelation)
		return
	}
	log.Printf("event=control_plane_request_failed route=%q stage=%q error_type=%q trace_id=%q span_id=%q request_id=%q", sseLogRoute(r), boundedRequestFailureStage(stage), boundedRequestFailureType(err), traceID, spanID, requestCorrelationID(r))
}

func boundedRequestFailureStage(stage string) string {
	switch stage {
	case "authorization", "projection_reconcile", "thread_history", "runtime_snapshot_decode", "runtime_snapshot_sync":
		return stage
	default:
		return "unknown"
	}
}

func boundedRequestFailureType(err error) string {
	switch {
	case err == nil:
		return "none"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, ErrAuthorizationForbidden):
		return "forbidden"
	case errors.Is(err, ErrAuthorizationUnauthenticated):
		return "unauthenticated"
	case errors.Is(err, ErrAuthorizationInvalidCredential):
		return "invalid_credential"
	case errors.Is(err, errAssignmentProjectionConflict):
		return "assignment_conflict"
	default:
		var timeout net.Error
		var syntax *json.SyntaxError
		var tooLarge *http.MaxBytesError
		var snapshotValidation runtimeSnapshotValidationError
		switch {
		case errors.As(err, &snapshotValidation):
			return "invalid_runtime_snapshot"
		case errors.As(err, &timeout) && timeout.Timeout():
			return "timeout"
		case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
			return "incomplete_body"
		case errors.As(err, &syntax):
			return "malformed_json"
		case errors.As(err, &tooLarge):
			return "request_too_large"
		default:
			return "unknown"
		}
	}
}

func requestTraceID(r *http.Request) string {
	traceID, _ := traceCorrelationFromContext(r.Context())
	return traceID
}

func requestSpanID(r *http.Request) string {
	_, spanID := traceCorrelationFromContext(r.Context())
	return spanID
}
