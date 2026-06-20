package main

import (
	"os"
	"strings"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type aiAgentSnapshotCadence struct {
	Reload        time.Duration
	HeartbeatSave time.Duration
}

func aiAgentClientDevelopmentFromEnv() (bool, error) {
	return envOptionalBool(envAIAgentClientDev)
}

func aiAgentSnapshotCadenceFromEnv() (aiAgentSnapshotCadence, error) {
	reload, err := envOptionalDurationSeconds(envAIAgentClientSnapshotReload)
	if err != nil {
		return aiAgentSnapshotCadence{}, err
	}
	heartbeatSave, err := envOptionalDurationSeconds(envAIAgentClientHeartbeatSave)
	if err != nil {
		return aiAgentSnapshotCadence{}, err
	}
	return aiAgentSnapshotCadence{Reload: reload, HeartbeatSave: heartbeatSave}, nil
}

func aiAgentClientSnapshotStoreFromEnv(enabled bool, metrics *riidoaiserver.AIAgentClientPersistenceMetrics) (riidoaiserver.AIAgentClientSnapshotStore, error) {
	if !enabled {
		return nil, nil
	}
	base, err := dynamoDBConfigForFeature(envAIAgentClientDev)
	if err != nil {
		return nil, err
	}
	store, err := riidoaiserver.NewDynamoDBAIAgentClientSnapshot(riidoaiserver.DynamoDBAIAgentClientSnapshotConfig{
		Region:              base.region,
		TableName:           strings.TrimSpace(os.Getenv(envAIAgentClientTable)),
		Endpoint:            base.endpoint,
		CredentialsProvider: base.provider,
		Metrics:             metrics,
	})
	if err != nil {
		return nil, wrapEnvError(envAIAgentClientTable, err)
	}
	return store, nil
}
