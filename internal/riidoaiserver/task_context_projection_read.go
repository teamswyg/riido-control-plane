package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
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
	var projections privateDocumentProjections
	if c.projectionBundles {
		var out struct {
			Structured privateDocumentProjections `json:"structured"`
			Rendered   privateDocumentProjections `json:"rendered"`
		}
		if err := c.contextGraphQL(ctx, bearer, taskContextProjectionQuery, team, id, &out); err != nil {
			outcome = taskContextReadOutcome(ctx, err)
			return nil
		}
		projections.JSON, projections.HTML = out.Structured.JSON, out.Rendered.HTML
	} else {
		var err error
		projections, err = c.readRESTProviderProjections(ctx, team, id, bearer)
		if err != nil {
			outcome = taskContextReadOutcome(ctx, err)
			return nil
		}
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

func (c *AIAgentPrivateTaskContextClient) readRESTProviderProjections(ctx context.Context, team, id, bearer string) (privateDocumentProjections, error) {
	var out privateDocumentProjections
	for _, format := range []string{"json", "html"} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(c.privateProviderDocumentEndpoint(team, id), "html")+format, nil)
		if err != nil {
			return out, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+bearer)
		if err = c.doPrivateTaskContextJSON(req, &out); err != nil {
			var failure taskContextReadError
			if format == "json" && errors.As(err, &failure) && failure.outcome == "missing" {
				continue
			}
			return out, err
		}
		if format == "json" && out.JSON != nil {
			if _, _, err := readableTaskContextJSON(*out.JSON); err == nil {
				return out, nil
			}
		}
	}
	return out, nil
}
