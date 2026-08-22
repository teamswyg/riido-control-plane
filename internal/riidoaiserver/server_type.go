package riidoaiserver

import (
	"net/http"
	"time"
)

type ServerConfig struct {
	Authorizer               RequestAuthorizer
	AgentCatalogStore        AgentCatalogStore
	AIAgentClient            AIAgentClientStore
	AIAgentProfileThumbnails AIAgentProfileThumbnailUploadService
	DeviceCredentials        DeviceCredentialStore
	Assignment               AssignmentStore
	TaskContext              AIAgentTaskContextReader
	ProviderStatus           ProviderStatusStore
	ProviderRead             ProviderStatusReader
	WebAllowedOrigins        []string
	HTTPTransactions         *HTTPTransactionMetrics
	TraceRecorder            TraceRecorder
	ControlPlaneGraphQL      http.Handler
	// LongPollMaxHold caps how long a daemon claim poll (PollRequest.WaitMs) may
	// be held open. Zero applies the default (25s). Must stay well under the ALB
	// idle timeout (60s default) and the http.Server write/idle timeouts (unset).
	LongPollMaxHold time.Duration
	// LongPollTick is the fallback re-evaluation interval during a held poll. It
	// bounds cross-instance discovery latency (an assignment queued on another
	// control-plane instance). Zero applies the default (2s).
	LongPollTick time.Duration
	// AIAgentGlobalReconcileMinInterval coalesces workspace-wide AI Agent
	// projection repair. Task-scoped action/read repair still runs per request.
	AIAgentGlobalReconcileMinInterval time.Duration
}

type Server struct {
	assignment               AssignmentStore
	agentCatalog             AgentCatalogStore
	aiAgent                  AIAgentClientStore
	aiAgentProfileThumbnails AIAgentProfileThumbnailUploadService
	daemonRuntime            AIAgentDaemonRuntimeStore
	taskContext              AIAgentTaskContextReader
	provider                 ProviderStatusStore
	providerRead             ProviderStatusReader
	devices                  DeviceCredentialStore
	aiAgentGlobalReconcile   *aiAgentGlobalReconcileGate
	config                   ServerConfig
}
