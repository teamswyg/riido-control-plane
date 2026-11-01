package riidoaiserver

import "net/http"

func cliAgentHTTPRoute(method, path string) string {
	target, ok := cliAgentReadPath(path)
	if method != http.MethodGet || !ok {
		return ""
	}
	base := "/v1/cli/workspaces/{workspace_id}/tasks/{task_id}/agents"
	if target.agent != "" {
		return base + "/{agent_id}"
	}
	return base
}
