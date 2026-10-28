package riidoaiserver

import "testing"

func TestCommentInvocationRecoversPublishedChildBeforeOfflineCheck(t *testing.T) {
	f := newCommentCallFixture(t)
	body := mentionHTML("agent-owned-codex") + `<p>Compare the plans and produce the cost table.</p>`
	if w := f.complete(t, body); w.Code != 200 {
		t.Fatal(w.Code)
	}
	// Simulate a crash after publishing the child but before saving its outcome.
	f.client.taskThreads[f.source.TaskID][0].CommentInvocation.Outcomes = nil
	for i := range f.client.devices {
		for j := range f.client.devices[i].Runtimes {
			f.client.devices[i].Runtimes[j].Availability = RuntimeAvailabilityOffline
		}
	}
	if err := f.client.saveSnapshot(t.Context()); err != nil {
		t.Fatal(err)
	}
	f.client = openPersistentAIAgentClientStoreForDelegateTest(t, f.snapshots)
	f.s.aiAgent = f.client
	if w := f.complete(t, body); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	rows := f.threads(t)
	if len(rows) != 2 || rows[0].CommentInvocation.Outcomes["target:agent-owned-codex"] != "accepted" {
		t.Fatalf("published child was lost on retry: %+v", rows)
	}
	for _, m := range f.client.taskThreadMessages[f.source.ThreadID] {
		if m.Role == "system" {
			t.Fatal("successful call misreported as a failure")
		}
	}
}
