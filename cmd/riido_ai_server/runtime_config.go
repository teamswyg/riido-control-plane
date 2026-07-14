package main

import (
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type runtimeConfig struct {
	Addr                     string
	ShutdownTimeout          time.Duration
	Authorizer               riidoaiserver.RequestAuthorizer
	ReviewProvision          *riidoaiserver.ReviewAccountProvisioning
	MetricsLogInterval       time.Duration
	PprofAddr                string
	Tracing                  tracingRuntimeConfig
	WebAllowedOrigins        []string
	AssignmentActiveLease    time.Duration
	LongPollMaxHold          time.Duration
	LongPollTick             time.Duration
	AIAgentClientDev         bool
	AIAgentClientStore       riidoaiserver.AIAgentClientSnapshotStore
	AIAgentClientMetrics     *riidoaiserver.AIAgentClientPersistenceMetrics
	AIAgentDaemonProfile     string
	DaemonClientPolicy       riidoaiserver.DaemonClientCompatibilityPolicy
	AIAgentSnapshotReload    time.Duration
	AIAgentHeartbeatSave     time.Duration
	AIAgentProfileThumbnails riidoaiserver.AIAgentProfileThumbnailUploadService
	AssignmentOperationStore riidoaiserver.AssignmentOperationStore
	AssignmentOutbox         riidoaiserver.EventSink
	TaskContextReader        riidoaiserver.AIAgentTaskContextReader
}

func closeRuntimeConfig(config runtimeConfig) {
	closeIfSupported(config.AIAgentClientStore)
	closeIfSupported(config.AssignmentOperationStore)
	closeIfSupported(config.AssignmentOutbox)
}

func closeIfSupported(value any) {
	if closer, ok := value.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}
