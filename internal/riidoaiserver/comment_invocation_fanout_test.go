package riidoaiserver

import "testing"

func TestCommentInvocationFanoutPreservesMentionOrderAndBudget(t *testing.T) {
	f := newCommentCallFixture(t)
	for i := range f.client.devices {
		for j := range f.client.devices[i].Runtimes {
			runtime := &f.client.devices[i].Runtimes[j]
			runtime.Availability = RuntimeAvailabilityOnline
			runtime.DetectionState = RuntimeDetectionStateDetected
		}
	}
	agent := f.client.agents["agent-owned-claude"]
	agent.WorkStatus = AgentWorkStatusIdle
	f.client.agents[agent.AgentID] = agent
	if err := f.client.saveSnapshot(t.Context()); err != nil {
		t.Fatal(err)
	}
	body := `<p>` + mentionHTML("agent-owned-claude") + mentionHTML("agent-owned-codex") + mentionHTML("agent-owned-claude") + ` Compare plans A and B by cost.</p>`
	if w := f.complete(t, body); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	rows := f.threads(t)
	if len(rows) != 3 {
		t.Fatalf("expected two actual children: %+v", rows)
	}
	for i, id := range []string{"agent-owned-claude", "agent-owned-codex"} {
		child := rows[i+1]
		if child.AgentID != id || child.CommentInvocation.Remaining != 3 {
			t.Fatalf("order/budget: %+v", child)
		}
		_, found, err := f.assignments.LoadAssignmentProjection(t.Context(), child.AssignmentID)
		if err != nil || !found {
			t.Fatal("missing real child assignment", err)
		}
	}
}
