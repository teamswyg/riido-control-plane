package riidoaiserver

import (
	"strings"
	"testing"
)

func BenchmarkSinglePassAssignmentPrompt(b *testing.B) {
	input := AIAgentAssignmentPromptInput{TaskID: "task", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task", Title: "Summarize this note"}, Document: AIAgentTaskContextDocument{Content: strings.Repeat("Bounded context. ", 300)}}}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := ComposeAIAgentAssignmentPrompt(input); err != nil {
			b.Fatal(err)
		}
	}
}
