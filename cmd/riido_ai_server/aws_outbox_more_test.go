package main

import (
	"strings"
	"testing"
)

func TestAWSCredentialsEndpointUsesRelativeURI(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAWSContainerCredentialsRelativeURI, "/v2/credentials")
	got, err := awsContainerCredentialsEndpointFromEnv("test feature")
	if err != nil {
		t.Fatalf("awsContainerCredentialsEndpointFromEnv: %v", err)
	}
	if !strings.HasSuffix(got, "/v2/credentials") {
		t.Fatalf("endpoint = %q, want ECS relative URI suffix", got)
	}
}

func TestAWSCredentialsEndpointRejectsRelativeURIWithoutSlash(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAWSContainerCredentialsRelativeURI, "v2/credentials")
	_, err := awsContainerCredentialsEndpointFromEnv("test feature")
	if err == nil || !strings.Contains(err.Error(), envAWSContainerCredentialsRelativeURI) {
		t.Fatalf("awsContainerCredentialsEndpointFromEnv err=%v", err)
	}
}

func TestAssignmentOutboxFromEnvNoopsWithoutTable(t *testing.T) {
	clearRiidoAIServerEnv(t)
	sink, err := assignmentOutboxFromEnv()
	if err != nil || sink != nil {
		t.Fatalf("assignmentOutboxFromEnv = %T, %v; want nil, nil", sink, err)
	}
}

func TestAssignmentOutboxFromEnvRejectsMissingRegion(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envDynamoDBOutboxTable, "outbox-table")
	_, err := assignmentOutboxFromEnv()
	if err == nil || !strings.Contains(err.Error(), envAWSRegion) {
		t.Fatalf("assignmentOutboxFromEnv err=%v", err)
	}
}

func TestAssignmentOutboxFromEnvBuildsSink(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envDynamoDBOutboxTable, "outbox-table")
	t.Setenv(envAWSRegion, "ap-northeast-2")
	t.Setenv(envAWSContainerCredentialsFullURI, "http://127.0.0.1/credentials")
	sink, err := assignmentOutboxFromEnv()
	if err != nil || sink == nil {
		t.Fatalf("assignmentOutboxFromEnv = %T, %v; want sink", sink, err)
	}
}
