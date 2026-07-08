package rendertext

import (
	"strings"
	"testing"
)

func TestLoopRendersAllSteps(t *testing.T) {
	var b strings.Builder
	Loop(&b, "obs", "hyp", "exec", "eval", "retro")
	got := b.String()
	for _, want := range []string{"| Observe | obs |", "| Execute | exec |", "| Retrospective | retro |"} {
		if !strings.Contains(got, want) {
			t.Fatalf("loop missing %q: %s", want, got)
		}
	}
}

func TestCodeListSkipsEmptyAndRendersValues(t *testing.T) {
	var empty strings.Builder
	CodeList(&empty, "Empty", nil)
	if empty.String() != "" {
		t.Fatalf("empty render = %q", empty.String())
	}
	var b strings.Builder
	CodeList(&b, "Paths", []string{"/v2/a", "/v3/b"})
	got := b.String()
	for _, want := range []string{"## Paths", "- `/v2/a`", "- `/v3/b`"} {
		if !strings.Contains(got, want) {
			t.Fatalf("code list missing %q: %s", want, got)
		}
	}
}
