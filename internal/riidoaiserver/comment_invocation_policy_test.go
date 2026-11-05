package riidoaiserver

import (
	"sync"
	"testing"
)

func TestCommentInvocationIgnoresHistoryAndChangedTerminalReplay(t *testing.T) {
	f := newCommentCallFixture(t)
	if w := f.complete(t, "Done without delegation"); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := f.complete(t, mentionHTML("agent-owned-codex")); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if len(f.threads(t)) != 1 {
		t.Fatal("changed terminal replay replaced committed final output")
	}
	e := f.assignments.completion
	e.Type = EventRiidoLog
	e.Message = mentionHTML("agent-owned-codex")
	if err := f.s.dispatchCommentMentions(t.Context(), e); err != nil {
		t.Fatal(err)
	}
	if len(f.threads(t)) != 1 {
		t.Fatal("imported/log mention executed")
	}
}

func TestCommentInvocationConcurrentReplayUsesOneReceipt(t *testing.T) {
	f := newCommentCallFixture(t)
	body := mentionHTML("agent-owned-codex") + `<p>Compare the plans using the original cost criterion.</p>`
	f.status = 503
	f.complete(t, body)
	f.status = 200
	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			if err := f.s.dispatchCommentMentions(t.Context(), f.assignments.completion); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if len(f.threads(t)) != 2 {
		t.Fatal("duplicate child after concurrent replay")
	}
}
