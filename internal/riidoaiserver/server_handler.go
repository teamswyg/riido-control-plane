package riidoaiserver

import "net/http"

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/v2/desktop/workspaces/", s.handleDesktopWorkspaceRoutes)
	mux.HandleFunc("/v2/client/workspaces/", s.handleAIAgentClientWorkspaceRoutes)
	mux.HandleFunc("/v1/client/ai-agent/bootstrap", s.handleAIAgentClientBootstrap)
	mux.HandleFunc("/v1/client/ai-agent/devices", s.handleAIAgentClientDevices)
	mux.HandleFunc("/v1/client/ai-agent/devices/", s.handleAIAgentClientDeviceRoutes)
	mux.HandleFunc("/v1/client/ai-agent/onboarding/fixtures", s.handleAIAgentClientOnboardingFixtures)
	mux.HandleFunc("/v1/client/ai-agent/onboarding/fixtures/", s.handleAIAgentClientOnboardingFixtures)
	mux.HandleFunc("/v1/client/ai-agent/profile-thumbnails/uploads", s.handleAIAgentClientProfileThumbnailUpload)
	mux.HandleFunc("/v1/client/ai-agent/tasks/", s.handleAIAgentClientTasks)
	mux.HandleFunc("/v1/client/ai-agent/agents", s.handleAIAgentClientAgents)
	mux.HandleFunc("/v1/client/ai-agent/agents/", s.handleAIAgentClientAgents)
	mux.HandleFunc("/v1/client/ai-agent/events", s.handleAIAgentClientEvents)
	mux.HandleFunc("/v1/daemon/runtime-snapshot", s.handleDaemonRuntimeSnapshot)
	mux.HandleFunc("/v1/daemon/agent-bindings", s.handleDaemonAgentBindings)
	mux.HandleFunc("/v1/agent-catalog", s.handleAgentCatalog)
	mux.HandleFunc("/v1/agent-catalog/", s.handleAgentCatalog)
	mux.HandleFunc("/v1/component-tasks/", s.handleComponentTasks)
	mux.HandleFunc("/v1/agents/", s.handleAgents)
	var handler http.Handler = mux
	if len(s.config.WebAllowedOrigins) > 0 {
		handler = s.withWebFrontendCORS(handler)
	}
	handler = withHTTPTransactionMetrics(handler, s.config.HTTPTransactions)
	handler = withHTTPTracing(handler, s.config.TraceRecorder)
	return handler
}
