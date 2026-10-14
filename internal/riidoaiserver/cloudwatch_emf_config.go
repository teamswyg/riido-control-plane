package riidoaiserver

import "io"

const (
	defaultCloudWatchNamespace   = "Riido/RiidoAIServer"
	defaultCloudWatchServiceName = "riido_ai_server"
)

type CloudWatchEMFConfig struct {
	Namespace   string
	ServiceName string
	Writer      io.Writer
}

func normalizeCloudWatchEMFConfig(config CloudWatchEMFConfig) CloudWatchEMFConfig {
	if config.Namespace == "" {
		config.Namespace = defaultCloudWatchNamespace
	}
	if config.ServiceName == "" {
		config.ServiceName = defaultCloudWatchServiceName
	}
	return config
}
