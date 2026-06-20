package main

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func openAIAgentClient(ctx context.Context, config runtimeConfig) (riidoaiserver.AIAgentClientStore, error) {
	if !config.AIAgentClientDev {
		return nil, nil
	}
	persistentClient, err := riidoaiserver.OpenPersistentAIAgentClientStore(ctx, riidoaiserver.NewDevelopmentAIAgentClientStore(), config.AIAgentClientStore)
	if err != nil {
		return nil, fmt.Errorf("open AI Agent development store: %w", err)
	}
	if err := persistentClient.ConfigureSnapshotCadence(config.AIAgentSnapshotReload, config.AIAgentHeartbeatSave); err != nil {
		return nil, fmt.Errorf("configure AI Agent development snapshot cadence: %w", err)
	}
	return persistentClient, nil
}

func openAssignmentStore(ctx context.Context, config runtimeConfig, client riidoaiserver.AIAgentClientStore, traceRecorder riidoaiserver.TraceRecorder) (*riidoaiserver.Store, error) {
	store, err := riidoaiserver.OpenStoreWithConfig(ctx, riidoaiserver.StoreConfig{
		AgentRegistry:       riidoaiserver.NewCompositeAgentRegistry(agentRegistryFromAIAgentClient(client)),
		ActiveLeaseDuration: config.AssignmentActiveLease,
		Outbox:              config.AssignmentOutbox,
		OperationStore:      config.AssignmentOperationStore,
		TraceRecorder:       traceRecorder,
	})
	if err != nil {
		return nil, fmt.Errorf("open assignment store: %w", err)
	}
	return store, nil
}

func agentRegistryFromAIAgentClient(client riidoaiserver.AIAgentClientStore) riidoaiserver.AgentRegistry {
	if registry, ok := client.(riidoaiserver.AgentRegistry); ok {
		return registry
	}
	return nil
}
