package riidoaiserver

import (
	"context"
	"testing"
)

func TestRecoveryInventoryRejectsWrongIdentityBeforeLookup(t *testing.T) {
	for _, change := range []string{"agent", "lease", "runtime"} {
		t.Run(change, func(t *testing.T) {
			store, ops, client, req := preflightFixture(t)
			agentID := ops.projection.Assignment.AgentID
			switch change {
			case "agent":
				agentID = "another-agent"
			case "lease":
				req.Metadata["lease_token"] = "wrong"
			case "runtime":
				req.RuntimeID = "another-runtime"
			}
			if _, err := store.RecordAgentEvent(context.Background(), agentID, req); err == nil {
				t.Fatal("wrong execution authorized")
			}
			select {
			case <-client.entered:
				t.Fatal("unauthorized request queried owner inventory")
			default:
			}
		})
	}
}
