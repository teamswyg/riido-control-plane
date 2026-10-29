package riidoaiserver

import "time"

func NewServer(config ServerConfig) Server {
	config = normalizeServerConfigNilInterfaces(config)
	if memory, ok := config.AssignmentContext.(*MemoryAssignmentContextStore); ok {
		memory.bindRetention(config.Assignment)
	}
	config.WebAllowedOrigins = normalizeWebAllowedOrigins(config.WebAllowedOrigins)
	if config.LongPollMaxHold <= 0 {
		config.LongPollMaxHold = 25 * time.Second
	}
	if config.LongPollTick <= 0 {
		config.LongPollTick = 2 * time.Second
	}
	if config.AIAgentGlobalReconcileMinInterval <= 0 {
		config.AIAgentGlobalReconcileMinInterval = time.Second
	}
	agentCatalog := config.AgentCatalogStore
	if agentCatalog == nil {
		if store, ok := config.Assignment.(AgentCatalogStore); ok {
			agentCatalog = store
		}
	}
	provider := config.ProviderStatus
	if provider == nil {
		if store, ok := config.Assignment.(ProviderStatusStore); ok {
			provider = store
		}
	}
	providerRead := config.ProviderRead
	if providerRead == nil {
		if reader, ok := provider.(ProviderStatusReader); ok {
			providerRead = reader
		}
	}
	devices := config.DeviceCredentials
	if devices == nil {
		if store, ok := config.AIAgentClient.(DeviceCredentialStore); ok {
			devices = store
		}
	}
	var daemonRuntime AIAgentDaemonRuntimeStore
	if store, ok := config.AIAgentClient.(AIAgentDaemonRuntimeStore); ok {
		daemonRuntime = store
	}
	executionGate := taskAgentExecutionGate(newTaskAgentExecutionGates())
	if owner, ok := config.Assignment.(interface{ sharedTaskAgentExecutionGate() taskAgentExecutionGate }); ok {
		executionGate = owner.sharedTaskAgentExecutionGate()
	}
	return Server{
		commentReactions:         &commentReactionDispatcher{jobs: map[string]*commentReactionJob{}, sent: map[string]string{}},
		assignment:               config.Assignment,
		agentCatalog:             agentCatalog,
		aiAgent:                  config.AIAgentClient,
		aiAgentProfileThumbnails: config.AIAgentProfileThumbnails,
		daemonRuntime:            daemonRuntime,
		taskContext:              config.TaskContext,
		provider:                 provider,
		providerRead:             providerRead,
		devices:                  devices,
		aiAgentGlobalReconcile:   newAIAgentGlobalReconcileGate(config.AIAgentGlobalReconcileMinInterval),
		taskAgentExecutionGate:   executionGate,
		config:                   config,
	}
}
