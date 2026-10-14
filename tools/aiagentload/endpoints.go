package main

import "strings"

func endpointsFor(cfg config) ([]endpoint, error) {
	switch cfg.Scenario {
	case "public":
		return publicEndpoints(), nil
	case "client-read":
		return clientReadEndpoints(cfg.WorkspaceID), nil
	default:
		return nil, errUnknownScenario(cfg.Scenario)
	}
}

func publicEndpoints() []endpoint {
	return []endpoint{
		{Method: "GET", Path: "/healthz"},
		{Method: "GET", Path: "/readyz"},
	}
}

func clientReadEndpoints(workspaceID string) []endpoint {
	ws := strings.TrimSpace(workspaceID)
	return []endpoint{
		{Method: "GET", Path: "/healthz"},
		{Method: "GET", Path: "/readyz"},
		{Method: "GET", Path: "/v2/client/workspaces/" + ws + "/ai-agent/bootstrap", Auth: true},
		{Method: "GET", Path: "/v2/client/workspaces/" + ws + "/ai-agent/devices", Auth: true},
		{Method: "GET", Path: "/v2/client/workspaces/" + ws + "/ai-agent/tasks/task-load-read/assignable-agents", Auth: true},
		{Method: "GET", Path: "/v2/client/workspaces/" + ws + "/ai-agent/tasks/task-load-read/threads", Auth: true},
		{Method: "GET", Path: "/v2/client/workspaces/" + ws + "/ai-agent/tasks/task-load-read/thread-stream-subscription", Auth: true},
	}
}
