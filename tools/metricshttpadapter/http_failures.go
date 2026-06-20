package main

import (
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func callMissingScope(m manifest) (int, error) {
	store := riidoaiserver.NewStore()
	defer store.Close()
	server, err := metricsServer(store, nil, scopedAuthorizer("limited-token", "agent:agent-a:poll"))
	if err != nil {
		return 0, err
	}
	status, _ := callMetrics(server, m, "limited-token")
	return status, nil
}

func callUnconfigured(m manifest) (int, error) {
	server, err := metricsServer(nil, nil, scopedAuthorizer("metrics-token", "metrics:read"))
	if err != nil {
		return 0, err
	}
	status, _ := callMetrics(server, m, "metrics-token")
	return status, nil
}
