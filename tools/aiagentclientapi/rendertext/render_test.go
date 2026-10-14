package rendertext

import (
	"strings"
	"testing"
)

func TestCodeListWrapsItems(t *testing.T) {
	got := CodeList([]string{"thread_id", "assignment_id"})
	want := "`thread_id`, `assignment_id`"
	if got != want {
		t.Fatalf("code list = %q, want %q", got, want)
	}
}

func TestListRendersSection(t *testing.T) {
	var b strings.Builder
	List(&b, "Fields", []string{"agent_id", "run_id"})
	got := b.String()
	for _, want := range []string{"## Fields", "- `agent_id`, `run_id`"} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered list missing %q: %s", want, got)
		}
	}
}
