package riidoaiserver

import (
	"context"
	"net/http"
	"sync"
	"testing"
)

type cliPolicyOwnerFunc func(context.Context, string) (string, error)

func (f cliPolicyOwnerFunc) cliInstallationDeviceOwner(ctx context.Context, id string) (string, error) {
	return f(ctx, id)
}

func TestCLIInstallationPolicyFencesClaimDuringRead(t *testing.T) {
	f := newCLIInstallationFixture(t)
	intent := f.intent(t)
	if w := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}, true); w.Code != http.StatusOK {
		t.Fatalf("confirm=%d", w.Code)
	}
	var once sync.Once
	owners := cliPolicyOwnerFunc(func(ctx context.Context, id string) (string, error) {
		once.Do(func() {
			_, err := f.actor.AssignTask(ctx, "task", AssignRequest{ComponentID: "component", AgentID: "agent", RuntimeProvider: "codex", Prompt: "fixture", CreatedBy: "requester"})
			if err != nil {
				t.Fatal(err)
			}
			claimCtx := context.WithValue(ctx, authenticatedClaimDeviceKey{}, authenticatedClaimDevice{id, "owner"})
			if _, err = f.actor.PollAgent(claimCtx, "agent", PollRequest{}); err != nil {
				t.Fatal(err)
			}
		})
		return "owner", nil
	})
	policy, err := f.actor.readCLIInstallationPolicy(t.Context(), f.request.cliInstallationIdentity, owners)
	if err != nil || policy.Decision != "blocked" || policy.ExecutionRevision != "2" {
		t.Fatalf("stale pre-claim policy escaped fence: %+v %v", policy, err)
	}
}
