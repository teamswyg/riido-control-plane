package riidoaiserver

import (
	"strings"
	"testing"
)

func TestCLIModePromptSummarizesParentsAndListsRelatedTitles(t *testing.T) {
	c := extendedAssignmentContextForTest(3)
	c.Hierarchy.Project.Document.Content = strings.Repeat("프로젝트 본문 ", 200)
	c.Extended.Comments[0].Priority = TaskContextCommentPriorityTrigger
	c.Extended.Comments[0].Author = "유래"
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: "task-private", Context: c, UserInstruction: "요약해줘"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"project reference summary", "riido work get RIID-4800", "API 계약 정리", "주간 회의", "- 유래: 댓글 0", "2 more comment(s) not included", "riido work comments RIID-5900"} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("CLI prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
	if strings.Count(prompt.Prompt, "프로젝트 본문") > 80 || strings.Contains(prompt.Prompt, "댓글 1") {
		t.Fatal("parent bodies and lower-priority comments belong to the CLI, not the prompt")
	}
	classic, _ := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: "task-private", Context: assignmentContextTestTaskContext()})
	if strings.Contains(classic.Prompt, "riido work") || !strings.Contains(classic.Prompt, "project body") {
		t.Fatal("without extended context the prompt must stay as before")
	}
}
