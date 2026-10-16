package riidoaiserver

import "strings"

const (
	taskIntentExplicit       = "explicit_instruction"
	taskIntentIntentOriented = "intent_oriented"
	taskIntentMetadataOnly   = "metadata_only"
)

func writePromptTaskInterpretation(
	builder *strings.Builder,
	intentClass string,
) {
	builder.WriteString("## Task Interpretation\n")
	writePromptLine(builder, "intent_class", intentClass)
	writePromptLine(builder, "intent_gate_required", "false")
	writePromptLine(builder, "first_response_policy", firstResponsePolicy(intentClass))
	writePromptLine(builder, "clarification_policy", "identify_the_specific_blocker_and_ask_one_focused_question_only_if_needed")
	writePromptLine(builder, "provider_limit_result_message", clientMessageCloudCreditInsufficient)
	builder.WriteString("\n")
}

func firstResponsePolicy(intentClass string) string {
	if intentClass == taskIntentExplicit {
		return "execute_the_explicit_instruction"
	}
	return "act_on_requested_outcome_ask_only_for_specific_blocking_information"
}
