package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func assignmentOperationStoreFromEnv(enabled bool, activeLeaseDuration time.Duration) (riidoaiserver.AssignmentOperationStore, error) {
	if !enabled {
		return nil, nil
	}
	base, err := dynamoDBConfigForFeature(envAIAgentClientDev)
	if err != nil {
		return nil, err
	}
	store, err := riidoaiserver.NewDynamoDBAssignmentOperationStore(riidoaiserver.DynamoDBAssignmentOperationStoreConfig{
		Region:              base.region,
		TableName:           strings.TrimSpace(os.Getenv(envAIAgentClientTable)),
		Endpoint:            base.endpoint,
		ActiveLeaseDuration: activeLeaseDuration,
		CredentialsProvider: base.provider,
	})
	if err != nil {
		return nil, fmt.Errorf("%s assignment operation store: %w", envAIAgentClientTable, err)
	}
	return store, nil
}
