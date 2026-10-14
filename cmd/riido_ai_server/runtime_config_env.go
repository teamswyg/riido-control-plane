package main

import (
	"path/filepath"

	"github.com/teamswyg/riido-control-plane/internal/agentstatefile"
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func configFromEnv() (runtimeConfig, error) {
	timing, err := runtimeTimingFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	reviewProvision, err := reviewAccountProvisioningFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	authorizer, err := authorizerFromEnvWithReview(reviewProvision)
	if err != nil {
		return runtimeConfig{}, err
	}
	tracing, pprofAddr, err := observabilityConfigFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	aiAgentClientDev, err := aiAgentClientDevelopmentFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	localDir, err := localStateDirectory(authorizer)
	if err != nil {
		return runtimeConfig{}, err
	}
	if localDir != "" {
		aiAgentClientDev = true
	}
	aiAgentClientMetrics := riidoaiserver.NewAIAgentClientPersistenceMetrics()
	var aiAgentClientStore riidoaiserver.AIAgentClientSnapshotStore
	if localDir != "" {
		aiAgentClientStore = agentstatefile.NewFileAIAgentClientSnapshot(filepath.Join(localDir, "agents.json"))
	} else {
		aiAgentClientStore, err = aiAgentClientSnapshotStoreFromEnv(aiAgentClientDev, aiAgentClientMetrics)
	}
	if err != nil {
		return runtimeConfig{}, err
	}
	return configFromEnvParts(timing, runtimeConfigParts{
		localDir:             localDir,
		reviewProvision:      reviewProvision,
		authorizer:           authorizer,
		tracing:              tracing,
		pprofAddr:            pprofAddr,
		aiAgentClientDev:     aiAgentClientDev,
		aiAgentClientStore:   aiAgentClientStore,
		aiAgentClientMetrics: aiAgentClientMetrics,
	})
}

type runtimeConfigParts struct {
	localDir             string
	reviewProvision      *riidoaiserver.ReviewAccountProvisioning
	authorizer           riidoaiserver.RequestAuthorizer
	tracing              tracingRuntimeConfig
	pprofAddr            string
	aiAgentClientDev     bool
	aiAgentClientStore   riidoaiserver.AIAgentClientSnapshotStore
	aiAgentClientMetrics *riidoaiserver.AIAgentClientPersistenceMetrics
}
