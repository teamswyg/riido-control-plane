package main

import (
	"context"
	"net/http"
	"net/http/httptest"

	srv "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const pressureTokenHeader = "X-Riido-Ai-Agent-Token"

func buildHTTPEndpointThreadsV3(cfg config) (pressureOperation, error) {
	store, principal, taskID, err := pressureFixture(cfg)
	if err != nil {
		return pressureOperation{}, err
	}
	authorizer, err := srv.NewStaticTokenAuthorizer([]srv.StaticTokenCredential{{
		PrincipalID: principal.PrincipalID,
		Token:       "pressure-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	if err != nil {
		return pressureOperation{}, err
	}
	handler := srv.NewServer(srv.ServerConfig{
		AIAgentClient:    store,
		Authorizer:       authorizer,
		HTTPTransactions: srv.NewHTTPTransactionMetrics(),
	}).Handler()
	path := "/v3/client/workspaces/" + fixtureWorkspaceID + "/ai-agent/tasks/" + taskID + "/threads"
	return newPressureOperation(func() error {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
		req.Header.Set(pressureTokenHeader, "pressure-token")
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			return errUnexpectedStatus(resp.Code)
		}
		return nil
	}), nil
}
