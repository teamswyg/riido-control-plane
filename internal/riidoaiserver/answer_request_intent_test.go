package riidoaiserver

import "testing"

func TestAnswerOnlyProseRemainsAdvisoryBackground(t *testing.T) {
	for _, prose := range []string{
		"This section explains the strategy.", "This section compares the alternatives.",
		"This paragraph summarizes the planning note.", "The glossary translates the marketing terms.",
		"A summary of this planning note", "A translation of this marketing plan",
		"An explanation of this planning note", "A comparison of these marketing plans",
		"이 문서는 전략 설명을 담는다.", "기획안 요약 및 마케팅 비교 자료",
	} {
		t.Run(prose, func(t *testing.T) {
			input := answerRequestIntentInput()
			input.Context.Document.Content = prose
			assertAdvisoryAnswerRequestClassification(t, input, true)
		})
	}
}

func TestNominalAnswerRequestsAreScopedToRequestFields(t *testing.T) {
	for _, request := range []string{"Give me a summary of this planning note", "A translation of this marketing plan", "An explanation of this planning note", "A comparison of these marketing plans", "Please provide a summary of this planning note", "Summary of this planning note", "Translation of this marketing plan"} {
		t.Run(request, func(t *testing.T) {
			input := answerRequestIntentInput()
			input.UserInstruction = request
			assertAdvisoryAnswerRequestClassification(t, input, false)
			input.UserInstruction = ""
			input.Context.Component.Title = request
			assertAdvisoryAnswerRequestClassification(t, input, false)
		})
	}
}

func TestAnswerOnlyDirectiveSyntaxRemainsExecutable(t *testing.T) {
	for _, request := range []string{"Summarize this planning note", "Please explain this marketing plan", "Could you translate this planning note?", "Compare these marketing plans", "이 기획안을 요약", "이 차이를 설명", "이 내용을 번역", "이 두 계획을 비교"} {
		input := answerRequestIntentInput()
		input.Context.Document.Content = "Planning context.\n" + request
		assertAdvisoryAnswerRequestClassification(t, input, false)
	}
}

func answerRequestIntentInput() AIAgentAssignmentPromptInput {
	return AIAgentAssignmentPromptInput{TaskID: "task", Context: AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{ID: "task", ComponentType: "task", Title: "Marketing planning"},
		Document:  AIAgentTaskContextDocument{Content: "Marketing strategy and audience goals."},
	}}
}

func assertAdvisoryAnswerRequestClassification(t *testing.T, input AIAgentAssignmentPromptInput, want bool) {
	t.Helper()
	got, err := ComposeAIAgentAssignmentPrompt(input)
	if err != nil {
		t.Fatal(err)
	}
	if got.IntentGateRequired {
		t.Fatalf("intent gate=%v want=%v for title=%q instruction=%q document=%q", got.IntentGateRequired, want, input.Context.Component.Title, input.UserInstruction, input.Context.Document.Content)
	}
	if advisoryExplicit := classifyTaskContextIntent(input.Context.Component, input.Context.Document, input.UserInstruction) == taskIntentExplicit; advisoryExplicit == want {
		t.Fatalf("advisory classification changed: explicit=%v background=%v", advisoryExplicit, want)
	}
}

func TestAnswerOnlyProseTitleRemainsAdvisoryBackground(t *testing.T) {
	for _, title := range []string{"This section explains the strategy", "This section compares the alternatives", "기획안 요약 및 마케팅 비교 자료"} {
		input := answerRequestIntentInput()
		input.Context.Component.Title = title
		assertAdvisoryAnswerRequestClassification(t, input, true)
	}
}
