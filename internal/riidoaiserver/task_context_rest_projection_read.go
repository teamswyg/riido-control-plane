package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

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
