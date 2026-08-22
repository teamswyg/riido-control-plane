package main

import "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

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
	controlPlaneGraphQLMTLS, err := controlPlaneGraphQLMTLSConfigFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	aiAgentClientDev, err := aiAgentClientDevelopmentFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	aiAgentClientMetrics := riidoaiserver.NewAIAgentClientPersistenceMetrics()
	aiAgentClientStore, err := aiAgentClientSnapshotStoreFromEnv(aiAgentClientDev, aiAgentClientMetrics)
	if err != nil {
		return runtimeConfig{}, err
	}
	return configFromEnvParts(timing, runtimeConfigParts{
		reviewProvision:         reviewProvision,
		authorizer:              authorizer,
		tracing:                 tracing,
		pprofAddr:               pprofAddr,
		controlPlaneGraphQLMTLS: controlPlaneGraphQLMTLS,
		aiAgentClientDev:        aiAgentClientDev,
		aiAgentClientStore:      aiAgentClientStore,
		aiAgentClientMetrics:    aiAgentClientMetrics,
	})
}

type runtimeConfigParts struct {
	reviewProvision         *riidoaiserver.ReviewAccountProvisioning
	authorizer              riidoaiserver.RequestAuthorizer
	tracing                 tracingRuntimeConfig
	pprofAddr               string
	controlPlaneGraphQLMTLS controlPlaneGraphQLMTLSConfig
	aiAgentClientDev        bool
	aiAgentClientStore      riidoaiserver.AIAgentClientSnapshotStore
	aiAgentClientMetrics    *riidoaiserver.AIAgentClientPersistenceMetrics
}
