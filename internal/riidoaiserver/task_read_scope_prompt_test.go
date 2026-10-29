package riidoaiserver

import (
	"strings"
	"testing"
)

func TestTaskReadScopePromptAdvertisesStoredPartContinuation(t *testing.T) {
	current := assignmentContextTestTaskContext()
	current.Extended = nil
	current.ReadScope = &AssignmentReadScope{SchemaVersion: taskReadScopeSchema, RootID: current.Component.ID, Status: "partial"}
	current.Document.Status = "partial"
	current.Hierarchy.Project.Document = nil
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: current.Component.ID, Context: current})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"riido context get", "riido context part <source-id> --cursor <opaque>", "riido work get", "riido meeting get", "riido agents list", "incomplete"} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Errorf("scoped prompt missing %q", want)
		}
	}
	if strings.Contains(prompt.Prompt, "Document unavailable: do not invent") {
		t.Fatal("uncopied parent body falsely called unavailable")
	}
}
