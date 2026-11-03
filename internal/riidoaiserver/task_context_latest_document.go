package riidoaiserver

import (
	"context"
	"strings"
)

type latestTaskDocumentError struct{ cause error }

func (e latestTaskDocumentError) Error() string {
	return "latest task document unavailable; retry the same request_id"
}
func (e latestTaskDocumentError) Unwrap() error { return e.cause }

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
