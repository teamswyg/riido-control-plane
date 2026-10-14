package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAIAgentPrivateTaskContextClientRejectsInterfaceWithoutBearer(t *testing.T) {
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{
		BaseURL: "https://riido.example.test",
	})
	if err != nil {
		t.Fatalf("NewAIAgentPrivateTaskContextClient: %v", err)
	}
	_, err = client.GetAIAgentTaskContext(context.Background(), "component-a")
	if err == nil || !strings.Contains(err.Error(), "request-scoped bearer token") {
		t.Fatalf("GetAIAgentTaskContext err=%v", err)
	}
}

func TestAIAgentPrivateTaskContextClientHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (*AIAgentPrivateTaskContextClient)(nil).GetAIAgentTaskContext(ctx, "component-a")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GetAIAgentTaskContext canceled err=%v", err)
	}
}
