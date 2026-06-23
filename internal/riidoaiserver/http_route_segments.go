package riidoaiserver

import "strings"

func httpRouteSegments(path string) []string {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

func aiAgentClientV2Route(path string) string {
	segments := httpRouteSegments(path)
	if len(segments) < 5 ||
		segments[0] != "v2" ||
		segments[1] != "client" ||
		segments[2] != "workspaces" ||
		segments[4] != "ai-agent" {
		return ""
	}
	base := "/v2/client/workspaces/{workspace_id}/ai-agent"
	return aiAgentClientRouteFromSegments(base, segments[5:])
}

func aiAgentClientV1Route(path string) string {
	const base = "/v1/client/ai-agent"
	if !strings.HasPrefix(path, base) {
		return ""
	}
	rest := strings.Trim(strings.TrimPrefix(path, base), "/")
	return aiAgentClientRouteFromSegments(base, httpRouteSegments(rest))
}
