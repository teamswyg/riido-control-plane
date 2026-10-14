package riidoaiserver

import (
	"strings"
	"testing"
)

func assertPromptHasAll(t *testing.T, prompt string, values []string) {
	t.Helper()
	for _, want := range values {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}
