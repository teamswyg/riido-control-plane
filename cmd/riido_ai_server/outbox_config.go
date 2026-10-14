package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func assignmentOutboxFromEnv() (riidoaiserver.EventSink, error) {
	tableName := strings.TrimSpace(os.Getenv(envDynamoDBOutboxTable))
	if tableName == "" {
		return nil, nil
	}
	region := strings.TrimSpace(os.Getenv(envAWSRegion))
	if region == "" {
		return nil, fmt.Errorf("%s is required when %s is configured", envAWSRegion, envDynamoDBOutboxTable)
	}
	provider, err := awsContainerCredentialsProviderFromEnv()
	if err != nil {
		return nil, err
	}
	outbox, err := riidoaiserver.NewDynamoDBOutbox(riidoaiserver.DynamoDBOutboxConfig{
		Region:              region,
		TableName:           tableName,
		Endpoint:            strings.TrimSpace(os.Getenv(envDynamoDBEndpoint)),
		CredentialsProvider: provider,
	})
	if err != nil {
		return nil, wrapEnvError(envDynamoDBOutboxTable, err)
	}
	return outbox, nil
}
