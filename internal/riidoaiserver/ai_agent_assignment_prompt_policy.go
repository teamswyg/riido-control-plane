package riidoaiserver

import "strings"

func writePromptInteractionPolicy(builder *strings.Builder) {
	builder.WriteString("## Interaction Policy\n")
	for _, rule := range aiAgentPromptInteractionPolicyRules() {
		builder.WriteString("- ")
		builder.WriteString(rule)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

func aiAgentPromptInteractionPolicyRules() []string {
	return []string{
		"Classify the task context as either an explicit instruction or background/intent before doing work.",
		"If the title or document is marketing, analysis, planning, or intent-oriented and the first concrete action is ambiguous, ask a concise clarification question in the existing AI Agent thread before producing deliverables.",
		"Use the user's apparent language and product tone when asking clarification questions or reporting provider limits.",
		"If document content is unavailable, infer only from task title, hierarchy, repository metadata, and user follow-up; when still unsure, ask what to do first.",
		"When a follow-up thread message supplies a concrete instruction, treat that latest user message as the current directive and use the latest task document context.",
		"If provider usage, quota, or cloud-credit limits stop the work, report the limit as the thread result and do not ask the user for local tool approval.",
	}
}
