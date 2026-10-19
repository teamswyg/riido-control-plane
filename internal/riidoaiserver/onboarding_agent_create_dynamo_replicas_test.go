package riidoaiserver

import (
	"sync"
	"testing"
)

func TestOnboardingAgentCreateDynamoConcurrentReplicasConverge(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	left, _ := newMembershipDynamoReplica(t, fixture)
	right, _ := newMembershipDynamoReplica(t, fixture)
	before := len(left.agents)
	ctx := withOnboardingAgentCreateRequestID(t.Context(), "concurrent-replica-create")
	type result struct {
		response AgentClientRecordResponse
		err      error
	}
	results := make(chan result, 2)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(2)
	for _, replica := range []*PersistentAIAgentClientStore{left, right} {
		go func() {
			ready.Done()
			<-start
			response, err := replica.CreateAIAgentFromOnboardingFixture(ctx, profileColorTestPrincipal(), "riido_pm", profileColorTestRequest("one creation across replicas"))
			results <- result{response: response, err: err}
		}()
	}
	ready.Wait()
	close(start)
	first, second := <-results, <-results
	if first.err != nil || second.err != nil {
		t.Fatalf("same-key replicas did not recover the committed receipt: %v / %v", first.err, second.err)
	}
	if first.response.Agent.AgentID != second.response.Agent.AgentID {
		t.Fatalf("concurrent replicas returned different agents: %q / %q", first.response.Agent.AgentID, second.response.Agent.AgentID)
	}
	fresh, _ := newMembershipDynamoReplica(t, fixture)
	if len(fresh.agents) != before+1 {
		t.Fatalf("durable agent count delta=%d, want exactly one", len(fresh.agents)-before)
	}
	replayed, err := fresh.CreateAIAgentFromOnboardingFixture(ctx, profileColorTestPrincipal(), "riido_pm", profileColorTestRequest("one creation across replicas"))
	if err != nil || replayed.Agent.AgentID != first.response.Agent.AgentID {
		t.Fatalf("fresh replica lost the committed creation receipt: %v", err)
	}
}
