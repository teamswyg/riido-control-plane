package main

import (
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
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
	var assignmentOperationStore riidoaiserver.AssignmentOperationStore
	var assignmentOutbox riidoaiserver.EventSink
	if parts.localDir == "" {
		assignmentOperationStore, err = assignmentOperationStoreFromEnv(parts.aiAgentClientDev, timing.AssignmentActiveLease)
		if err != nil {
			return runtimeConfig{}, err
		}
		assignmentOutbox, err = assignmentOutboxFromEnv()
		if err != nil {
			return runtimeConfig{}, err
		}
	}
	profileThumbnails, err := agentProfileThumbnailUploadServiceFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	var snapshotStore riidoaiserver.SnapshotStore
	if parts.localDir != "" {
		snapshotStore, err = riidoaiserver.NewFileStoreSnapshot(filepath.Join(parts.localDir, "assignments.json"))
		if err != nil {
			return runtimeConfig{}, err
		}
	}
	taskContextReader, err := taskContextReaderFromEnv()
	if err == nil && parts.localDir != "" && taskContextReader == nil {
		taskContextReader, err = readLocalTaskContexts(parts.localDir)
	}
	if err != nil {
		return runtimeConfig{}, err
	}
	assignmentContext, err := assignmentContextStoreFromEnv(assignmentOperationStore)
	if err != nil {
		return runtimeConfig{}, err
	}
	commentDelivery := riidoaiserver.CommentDeliveryConfig{EvidenceBaseURL: os.Getenv("RIIDO_AI_SERVER_COMMENT_DELIVERY_EVIDENCE_URL"), Key: os.Getenv("RIIDO_AI_SERVER_COMMENT_DELIVERY_KEY")}
	if err := commentDelivery.Validate(); err != nil {
		return runtimeConfig{}, err
	}
	installationService, err := cliInstallationServiceFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	return runtimeConfig{
		CLIInstallationService:  installationService,
		CommentDelivery:         commentDelivery,
		AssignmentSnapshotStore: snapshotStore,

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
		AssignmentContext:        assignmentContext,
		TaskContextReader:        taskContextReader,
	}, nil
}
