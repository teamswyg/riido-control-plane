package main

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestOpenAIAgentClientDisabledReturnsNil(t *testing.T) {
	client, err := openAIAgentClient(context.Background(), runtimeConfig{})
	if err != nil {
		t.Fatalf("openAIAgentClient disabled: %v", err)
	}
	if client != nil {
		t.Fatalf("client = %T, want nil", client)
	}
}

func TestAgentRegistryFromAIAgentClient(t *testing.T) {
	if got := agentRegistryFromAIAgentClient(nil); got != nil {
		t.Fatalf("nil client registry = %T, want nil", got)
	}
	store := riidoaiserver.NewDevelopmentAIAgentClientStore()
	if got := agentRegistryFromAIAgentClient(store); got == nil {
		t.Fatal("development AI agent client should expose an agent registry")
	}
}

func TestOpenAssignmentStoreReturnsStore(t *testing.T) {
	store, err := openAssignmentStore(context.Background(), runtimeConfig{
		AssignmentActiveLease: time.Minute,
	}, riidoaiserver.NewDevelopmentAIAgentClientStore(), nil)
	if err != nil {
		t.Fatalf("openAssignmentStore: %v", err)
	}
	if store == nil {
		t.Fatal("openAssignmentStore returned nil store")
	}
}

func TestApplyReviewAccountProvisioningNoopWhenDisabled(t *testing.T) {
	if err := applyReviewAccountProvisioning(context.Background(), nil, runtimeConfig{}); err != nil {
		t.Fatalf("applyReviewAccountProvisioning disabled: %v", err)
	}
}
