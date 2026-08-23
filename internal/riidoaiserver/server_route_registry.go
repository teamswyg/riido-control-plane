package riidoaiserver

import "net/http"

type serverRoute struct {
	pattern string
	handler func(Server, http.ResponseWriter, *http.Request)
}

func registerServerRoutes(mux *http.ServeMux, s Server) {
	for _, route := range serverRoutes {
		mux.HandleFunc(route.pattern, func(w http.ResponseWriter, r *http.Request) {
			route.handler(s, w, r)
		})
	}
}

func (s Server) handleControlPlaneOwnerGraphQL(w http.ResponseWriter, r *http.Request) {
	if s.config.ControlPlaneOwnerGraphQL == nil {
		writeError(w, http.StatusServiceUnavailable, "control-plane owner GraphQL is not configured")
		return
	}
	s.config.ControlPlaneOwnerGraphQL.ServeHTTP(w, r)
}

var serverRoutes = []serverRoute{
	{"/healthz", Server.handleHealth},
	{"/readyz", Server.handleReady},
	{"/owner/graphql", Server.handleControlPlaneOwnerGraphQL},
	{"/metrics", Server.handleMetrics},
	{"/v2/desktop/workspaces/", Server.handleDesktopWorkspaceRoutes},
	{"/v2/client/workspaces/", Server.handleAIAgentClientWorkspaceRoutes},
	{"/v3/client/workspaces/", Server.handleAIAgentClientWorkspaceRoutes},
	{"/v1/client/ai-agent/bootstrap", Server.handleAIAgentClientBootstrap},
	{"/v1/client/ai-agent/devices", Server.handleAIAgentClientDevices},
	{"/v1/client/ai-agent/devices/", Server.handleAIAgentClientDeviceRoutes},
	{"/v1/client/ai-agent/onboarding/fixtures", Server.handleAIAgentClientOnboardingFixtures},
	{"/v1/client/ai-agent/onboarding/fixtures/", Server.handleAIAgentClientOnboardingFixtures},
	{"/v1/client/ai-agent/profile-thumbnails/uploads", Server.handleAIAgentClientProfileThumbnailUpload},
	{"/v1/client/ai-agent/agent-assignments/", Server.handleAIAgentClientAgentAssignments},
	{"/v1/client/ai-agent/tasks/", Server.handleAIAgentClientTasks},
	{"/v1/client/ai-agent/threads/", Server.handleAIAgentClientThreads},
	{"/v1/client/ai-agent/agents", Server.handleAIAgentClientAgents},
	{"/v1/client/ai-agent/agents/", Server.handleAIAgentClientAgents},
	{"/v1/client/ai-agent/events", Server.handleAIAgentClientEvents},
	{"/v1/daemon/runtime-snapshot", Server.handleDaemonRuntimeSnapshot},
	{"/v1/daemon/agent-bindings", Server.handleDaemonAgentBindings},
	{"/v1/agent-catalog", Server.handleAgentCatalog},
	{"/v1/agent-catalog/", Server.handleAgentCatalog},
	{"/v1/component-tasks/", Server.handleComponentTasks},
	{"/v1/agents/", Server.handleAgents},
}
