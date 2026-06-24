package riidoaiserver

import "strings"

const (
	taskIntentExplicit       = "explicit_instruction"
	taskIntentIntentOriented = "intent_oriented"
	taskIntentMetadataOnly   = "metadata_only"
)

func writePromptTaskInterpretation(
	builder *strings.Builder,
	component AIAgentTaskContextComponent,
	document AIAgentTaskContextDocument,
) {
	intentClass := classifyTaskContextIntent(component, document)
	builder.WriteString("## Task Interpretation\n")
	writePromptLine(builder, "intent_class", intentClass)
	writePromptLine(builder, "first_response_policy", firstResponsePolicy(intentClass))
	writePromptLine(builder, "clarification_question_example", clarificationQuestionExample(component))
	writePromptLine(builder, "provider_limit_result_message", clientMessageCloudCreditInsufficient)
	builder.WriteString("\n")
}

func classifyTaskContextIntent(
	component AIAgentTaskContextComponent,
	document AIAgentTaskContextDocument,
) string {
	text := strings.ToLower(component.Title + "\n" + document.Content)
	if strings.TrimSpace(document.Content) == "" {
		return taskIntentMetadataOnly
	}
	if containsAny(text, intentOrientedTaskMarkers()) {
		return taskIntentIntentOriented
	}
	return taskIntentExplicit
}

func firstResponsePolicy(intentClass string) string {
	switch intentClass {
	case taskIntentIntentOriented:
		return "ask_for_intent_before_deliverables_when_first_action_is_ambiguous"
	case taskIntentMetadataOnly:
		return "infer_from_metadata_then_ask_when_unsure"
	default:
		return "execute_the_explicit_instruction"
	}
}

func clarificationQuestionExample(component AIAgentTaskContextComponent) string {
	if looksKorean(component.Title) {
		return "어떤 작업부터 진행할까요?"
	}
	return "What should I work on first?"
}
