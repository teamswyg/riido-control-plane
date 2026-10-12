package main

import (
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"os"
)

func configFromEnvParts(timing runtimeTimingConfig, parts runtimeConfigParts) (runtimeConfig, error) {
	webAllowedOrigins, err := webAllowedOriginsFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	cadence, err := aiAgentSnapshotCadenceFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	daemonProfile, err := daemonProfileFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	assignmentOperationStore, err := assignmentOperationStoreFromEnv(parts.aiAgentClientDev, timing.AssignmentActiveLease)
	if err != nil {
		return runtimeConfig{}, err
	}
	assignmentOutbox, err := assignmentOutboxFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	profileThumbnails, err := agentProfileThumbnailUploadServiceFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	taskContextReader, err := taskContextReaderFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	commentDelivery := riidoaiserver.CommentDeliveryConfig{EvidenceBaseURL: os.Getenv("RIIDO_AI_SERVER_COMMENT_DELIVERY_EVIDENCE_URL"), Key: os.Getenv("RIIDO_AI_SERVER_COMMENT_DELIVERY_KEY")}
	if err := commentDelivery.Validate(); err != nil {
		return runtimeConfig{}, err
	}
	return runtimeConfig{
		CommentDelivery:          commentDelivery,
		Addr:                     getenvDefault(envAddr, ":8080"),
		ShutdownTimeout:          timing.ShutdownTimeout,
		Authorizer:               parts.authorizer,
		ReviewProvision:          parts.reviewProvision,
		MetricsLogInterval:       timing.MetricsLogInterval,
		PprofAddr:                parts.pprofAddr,
		Tracing:                  parts.tracing,
		WebAllowedOrigins:        webAllowedOrigins,
		AssignmentActiveLease:    timing.AssignmentActiveLease,
		LongPollMaxHold:          timing.LongPollMaxHold,
		LongPollTick:             timing.LongPollTick,
		AIAgentClientDev:         parts.aiAgentClientDev,
		AIAgentClientStore:       parts.aiAgentClientStore,
		AIAgentClientMetrics:     parts.aiAgentClientMetrics,
		AIAgentDaemonProfile:     daemonProfile,
		DaemonClientPolicy:       daemonClientCompatibilityPolicyFromEnv(),
		AIAgentSnapshotReload:    cadence.Reload,
		AIAgentHeartbeatSave:     cadence.HeartbeatSave,
		AIAgentProfileThumbnails: profileThumbnails,
		AssignmentOperationStore: assignmentOperationStore,
		AssignmentOutbox:         assignmentOutbox,
		TaskContextReader:        taskContextReader,
	}, nil
}
