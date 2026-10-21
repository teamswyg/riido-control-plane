package riidoaiserver

import (
	"context"
)

const taskContextProjectionQuery = `query AgentTaskProjection($id: String!, $team: String!) {
 structured: getProviderDocument(id:$id,teamId:$team,getProviderDocumentDto:{format:"json"})
 rendered: getProviderDocument(id:$id,teamId:$team,getProviderDocumentDto:{format:"html"})
}`

func (c *AIAgentPrivateTaskContextClient) readProviderProjection(ctx context.Context, team, id, bearer string) *AIAgentTaskContextDocument {
	ctx, finish := observeTaskContextRead(ctx, "document_projection")
	outcome, source, jsonState := "unavailable", "none", "missing"
	defer func() {
		finish(outcome, StringTraceAttribute("riido.task_context.source", source), StringTraceAttribute("riido.task_context.json_state", jsonState))
	}()
	projections, err := c.fetchProviderProjections(ctx, team, id, bearer)
	if err != nil {
		outcome = taskContextReadOutcome(ctx, err)
		return nil
	}
	if projections.JSON != nil {
		jsonState = "invalid"
	}
	document := projections.document()
	if document != nil {
		outcome, source = document.Status, document.Source
		if source == "provider_json_projection" {
			jsonState = outcome
		}
	}
	return document
}
