package riidoaiserver

import "context"

func (c *AIAgentPrivateTaskContextClient) fetchProviderProjections(ctx context.Context, team, id, bearer string) (privateDocumentProjections, error) {
	var projections privateDocumentProjections
	if c.projectionBundles {
		var out struct {
			Structured privateDocumentProjections `json:"structured"`
			Rendered   privateDocumentProjections `json:"rendered"`
		}
		if err := c.contextGraphQL(ctx, bearer, taskContextProjectionQuery, team, id, &out); err != nil {
			return projections, err
		}
		projections.JSON, projections.HTML = out.Structured.JSON, out.Rendered.HTML
	} else {
		var err error
		projections, err = c.readRESTProviderProjections(ctx, team, id, bearer)
		if err != nil {
			return projections, err
		}
	}
	return projections, nil
}
