package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProviderDocumentDrivesExplicitIntentClassification(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/public/components/component-a/workspace":
			writeProviderDocumentWorkspace(w)
		case "/teams/team-a/components/component-a":
			writeProviderDocumentComponent(w)
		case "/documents/providers/team-a/component-a":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"html": "<p>main.go를 만들고 Bye World를 출력해줘.</p>",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	contextSnapshot := readProviderDocumentContextFixture(t, server.URL)
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID:  "component-a",
		Context: contextSnapshot,
	})
	if err != nil {
		t.Fatalf("ComposeAIAgentAssignmentPrompt: %v", err)
	}
	for _, want := range []string{
		"- intent_class: explicit_instruction",
		"- intent_gate_required: false",
		"main.go를 만들고 Bye World를 출력해줘.",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}
