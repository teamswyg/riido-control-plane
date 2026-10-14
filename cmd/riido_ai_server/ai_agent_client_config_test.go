package main

import (
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestConfigFromEnvParsesAIAgentClientDevelopmentStore(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAIAgentClientDev, "true")
	t.Setenv(envAIAgentClientTable, "riido-ai-agent-development")
	t.Setenv(envAWSRegion, "ap-northeast-2")
	t.Setenv(envAWSContainerCredentialsFullURI, "http://169.254.170.2/credentials")
	t.Setenv(envAssignmentActiveLease, "300")
	t.Setenv(envAIAgentClientSnapshotReload, "17")
	t.Setenv(envAIAgentClientHeartbeatSave, "19")
	t.Setenv(envAIAgentDaemonProfile, "testnet")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	defer closeRuntimeConfig(config)
	if !config.AIAgentClientDev || config.AssignmentActiveLease != 5*time.Minute {
		t.Fatalf("AI Agent development timing config = %+v", config)
	}
	if config.AIAgentSnapshotReload != 17*time.Second || config.AIAgentHeartbeatSave != 19*time.Second {
		t.Fatalf("AI Agent snapshot cadence = reload:%s heartbeat:%s", config.AIAgentSnapshotReload, config.AIAgentHeartbeatSave)
	}
	if config.AIAgentDaemonProfile != "staging" {
		t.Fatalf("daemon profile = %q, want staging", config.AIAgentDaemonProfile)
	}
	if config.AIAgentClientStore == nil || config.AIAgentClientMetrics == nil || config.AssignmentOperationStore == nil {
		t.Fatalf("AI Agent persistence config missing: %+v", config)
	}
	assertAssignmentOperationStoreCapabilities(t, config.AssignmentOperationStore)
}

func assertAssignmentOperationStoreCapabilities(t *testing.T, store riidoaiserver.AssignmentOperationStore) {
	t.Helper()
	if _, ok := store.(riidoaiserver.AssignmentOperationLoader); !ok {
		t.Fatalf("assignment operation store should load operation journal, got %T", store)
	}
	if _, ok := store.(riidoaiserver.AssignmentClaimer); !ok {
		t.Fatalf("assignment operation store should claim queued assignments, got %T", store)
	}
	if _, ok := store.(riidoaiserver.AssignmentActiveLeaseStore); !ok {
		t.Fatalf("assignment operation store should persist active leases, got %T", store)
	}
}

func TestConfigFromEnvRejectsAIAgentClientDevelopmentWithoutDynamoDBTable(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAIAgentClientDev, "true")
	t.Setenv(envAWSRegion, "ap-northeast-2")
	t.Setenv(envAWSContainerCredentialsFullURI, "http://169.254.170.2/credentials")
	if _, err := configFromEnv(); err == nil || !strings.Contains(err.Error(), envAIAgentClientTable) {
		t.Fatalf("configFromEnv err=%v", err)
	}
}

func TestConfigFromEnvRejectsAIAgentClientDevelopmentWithoutCredentialEndpoint(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAIAgentClientDev, "true")
	t.Setenv(envAIAgentClientTable, "riido-ai-agent-development")
	t.Setenv(envAWSRegion, "ap-northeast-2")
	if _, err := configFromEnv(); err == nil || !strings.Contains(err.Error(), envAWSContainerCredentialsFullURI) {
		t.Fatalf("configFromEnv err=%v", err)
	}
}
