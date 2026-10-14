package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleAIAgentClientWorkspaceRoutes(w http.ResponseWriter, r *http.Request) {
	workspaceID, v1Path, ok := splitAIAgentClientWorkspacePath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	r = requestWithAIAgentWorkspaceIDAndPath(r, workspaceID, v1Path)
	switch {
	case strings.HasPrefix(v1Path, "/v3/client/ai-agent/tasks/"):
		s.handleAIAgentClientTasksV3(w, r)
	case v1Path == "/v1/client/ai-agent/bootstrap":
		s.handleAIAgentClientBootstrap(w, r)
	case v1Path == "/v1/client/ai-agent/devices":
		s.handleAIAgentClientDevices(w, r)
	case strings.HasPrefix(v1Path, "/v1/client/ai-agent/devices/"):
		s.handleAIAgentClientDeviceRoutes(w, r)
	case v1Path == "/v1/client/ai-agent/onboarding/fixtures" || strings.HasPrefix(v1Path, "/v1/client/ai-agent/onboarding/fixtures/"):
		s.handleAIAgentClientOnboardingFixtures(w, r)
	case v1Path == "/v1/client/ai-agent/profile-thumbnails/uploads":
		s.handleAIAgentClientProfileThumbnailUpload(w, r)
	case v1Path == "/v1/client/ai-agent/tasks/assigned-agent-profiles":
		s.handleAIAgentClientWorkspaceAssignedAgentProfiles(w, r)
	case strings.HasPrefix(v1Path, "/v1/client/ai-agent/agent-assignments/"):
		s.handleAIAgentClientAgentAssignments(w, r)
	case strings.HasPrefix(v1Path, "/v1/client/ai-agent/tasks/"):
		s.handleAIAgentClientTasks(w, r)
	case strings.HasPrefix(v1Path, "/v1/client/ai-agent/threads/"):
		s.handleAIAgentClientThreads(w, r)
	case v1Path == "/v1/client/ai-agent/agents" || strings.HasPrefix(v1Path, "/v1/client/ai-agent/agents/"):
		s.handleAIAgentClientAgents(w, r)
	case v1Path == "/v1/client/ai-agent/events":
		s.handleAIAgentClientEvents(w, r)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}
