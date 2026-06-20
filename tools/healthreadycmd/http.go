package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func verifyEndpoints(endpoints []endpointContract) ([]endpointEvidence, error) {
	handler := riidoaiserver.NewServer(riidoaiserver.ServerConfig{}).Handler()
	results := make([]endpointEvidence, 0, len(endpoints))
	for _, endpoint := range endpoints {
		req := httptest.NewRequest(endpoint.Method, endpoint.Path, nil)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		var health riidoaiserver.Health
		if err := json.Unmarshal(resp.Body.Bytes(), &health); err != nil {
			return nil, fmt.Errorf("decode %s: %w", endpoint.Path, err)
		}
		results = append(results, endpointEvidence{
			Name: endpoint.Name, Method: endpoint.Method, Path: endpoint.Path,
			HTTPStatus: resp.Code, Status: health.Status,
		})
	}
	return results, nil
}
