package main

import (
	"strings"
	"testing"
)

// Completion changes the canonical thread projection. Its invalidation must
// reach the existing v3 read key, not only the legacy v2 thread list.
func TestOnboardingCompletionInvalidatesExistingV3ThreadHistory(t *testing.T) {
	core, err := generate(loadTestOpenAPI(t))
	if err != nil {
		t.Fatal(err)
	}
	text := string(core)
	start := strings.LastIndex(text, "onboardingCompletion: {")
	if start < 0 {
		t.Fatal("completion mutation missing")
	}
	all := strings.Index(text[start:], "all: (queryClient: QueryClient) => Promise.all(")
	if all < 0 {
		t.Fatal("completion invalidation helper missing")
	}
	line := strings.SplitN(text[start+all:], "\n", 2)[0]
	if !strings.Contains(line, "listAIAgentTaskThreadHistoryV3QueryKeyRoot()") || !strings.Contains(text, "export async function listAIAgentTaskThreadHistoryV3") {
		t.Fatal("completion cannot invalidate the generated canonical history read")
	}
}
