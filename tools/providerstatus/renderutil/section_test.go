package renderutil

import (
	"strings"
	"testing"
)

func TestSection(t *testing.T) {
	var b strings.Builder
	Section(&b, "Providers", []string{"claude", "codex"})
	got := b.String()
	for _, want := range []string{"## Providers", "- `claude`", "- `codex`"} {
		if !strings.Contains(got, want) {
			t.Fatalf("section missing %q in %q", want, got)
		}
	}
}
