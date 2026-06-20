package riidoaiserver

func normalizeServerConfigNilInterfaces(config ServerConfig) ServerConfig {
	if isNilInterface(config.Authorizer) {
		config.Authorizer = nil
	}
	if isNilInterface(config.AgentCatalogStore) {
		config.AgentCatalogStore = nil
	}
	if isNilInterface(config.AIAgentClient) {
		config.AIAgentClient = nil
	}
	if isNilInterface(config.AIAgentProfileThumbnails) {
		config.AIAgentProfileThumbnails = nil
	}
	if isNilInterface(config.DeviceCredentials) {
		config.DeviceCredentials = nil
	}
	if isNilInterface(config.Assignment) {
		config.Assignment = nil
	}
	if isNilInterface(config.TaskContext) {
		config.TaskContext = nil
	}
	if isNilInterface(config.ProviderStatus) {
		config.ProviderStatus = nil
	}
	if isNilInterface(config.ProviderRead) {
		config.ProviderRead = nil
	}
	if isNilInterface(config.TraceRecorder) {
		config.TraceRecorder = nil
	}
	return config
}
