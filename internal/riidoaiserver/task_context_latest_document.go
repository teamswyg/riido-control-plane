package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type latestTaskDocumentError struct{ cause error }

func (e latestTaskDocumentError) Error() string {
	return "latest task document unavailable; retry the same request_id"
}
func (e latestTaskDocumentError) Unwrap() error { return e.cause }

// readScopeIntegrityReason is the monolith fault reason for task data that
// cannot be served safely (RIID-6008).
const readScopeIntegrityReason = "read_scope_integrity"

// taskReadScopeError marks a failure of the requester-authorized read scope
// itself, as opposed to other task-context reads (RIID-6008).
type taskReadScopeError struct{ cause error }

func (e taskReadScopeError) Error() string { return e.cause.Error() }

func (e taskReadScopeError) Unwrap() error { return e.cause }

// response keeps the retryable 503 for transient failures. A read-scope denial,
// missing task or integrity failure is reported with a non-retryable status
// because retrying the same request cannot fix it (RIID-6008). Other
// task-context reads keep the existing 503.
func (e latestTaskDocumentError) response() (int, string) {
	var scope taskReadScopeError
	var read taskContextReadError
	if errors.As(e.cause, &scope) && errors.As(scope.cause, &read) {
		switch read.outcome {
		case "denied":
			return http.StatusForbidden, "task read scope denied"
		case "missing":
			return http.StatusNotFound, "task not found"
		case "integrity":
			return http.StatusUnprocessableEntity, "task data cannot be read safely"
		}
	}
	return http.StatusServiceUnavailable, e.Error()
}

func (c *AIAgentPrivateTaskContextClient) getLatestProviderDocument(ctx context.Context, team, id, bearer string) (*AIAgentTaskContextDocument, error) {
	ctx, finish := observeTaskContextRead(ctx, "document_projection")
	outcome, source, jsonState := "unavailable", "none", "missing"
	defer func() {
		finish(outcome, StringTraceAttribute("riido.task_context.source", source), StringTraceAttribute("riido.task_context.json_state", jsonState))
	}()
	projections, err := c.fetchProviderProjections(ctx, team, id, bearer)
	if err != nil {
		outcome = taskContextReadOutcome(ctx, err)
		return nil, latestTaskDocumentError{err}
	}
	if projections.JSON != nil {
		jsonState = "invalid"
	}
	if projections.JSON != nil && strings.TrimSpace(*projections.JSON) != "" {
		if _, _, err := readableTaskContextJSON(*projections.JSON); err != nil {
			return nil, latestTaskDocumentError{err}
		}
	}
	document := projections.document()
	if document == nil {
		return nil, nil
	}
	outcome, source = document.Status, document.Source
	if source == "provider_json_projection" {
		jsonState = outcome
	}
	return document, nil
}
