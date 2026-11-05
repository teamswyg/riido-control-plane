package riidoaiserver

import (
	"testing"
)

func TestCommentEvidenceOutageKeepsTerminalEventPending(t *testing.T) {
	f := newCommentCallFixture(t)
	f.status = 503
	body := mentionHTML("agent-owned-codex") + `<p>Compare the plans and produce the cost table.</p>`
	if w := f.complete(t, body); w.Code != 503 {
		t.Fatalf("lost handoff: %d", w.Code)
	}
	if len(f.threads(t)) != 1 {
		t.Fatal("outage created a child")
	}
	f.status = 200
	if w := f.complete(t, body); w.Code != 200 {
		t.Fatalf("retry: %d %s", w.Code, w.Body.String())
	}
	if len(f.threads(t)) != 2 {
		t.Fatal("retry did not dispatch")
	}
}
