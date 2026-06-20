package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
)

type aiAgentWorkspaceIDContextKey struct{}

func aiAgentWorkspaceIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if value, ok := r.Context().Value(aiAgentWorkspaceIDContextKey{}).(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func requestWithAIAgentWorkspaceIDAndPath(r *http.Request, workspaceID, path string) *http.Request {
	urlCopy := *r.URL
	urlCopy.Path = path
	next := r.Clone(context.WithValue(r.Context(), aiAgentWorkspaceIDContextKey{}, strings.TrimSpace(workspaceID)))
	next.URL = &urlCopy
	return next
}

func splitAIAgentClientWorkspacePath(path string) (string, string, bool) {
	workspaceID, suffix, ok := splitNestedResourcePath(path, "/v2/client/workspaces/")
	if !ok || strings.TrimSpace(workspaceID) == "" {
		return "", "", false
	}
	suffix = strings.Trim(suffix, "/")
	switch {
	case suffix == "ai-agent":
		return workspaceID, "/v1/client/ai-agent", true
	case strings.HasPrefix(suffix, "ai-agent/"):
		return workspaceID, "/v1/client/ai-agent/" + strings.TrimPrefix(suffix, "ai-agent/"), true
	default:
		return "", "", false
	}
}
