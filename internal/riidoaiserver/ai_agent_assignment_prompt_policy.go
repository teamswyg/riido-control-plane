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
		"Interpret the requested outcome within this same response; do not start a separate planning, classification or review run.",
		"For questions, summaries, translations and explanations, answer directly from sufficient supplied context; do not require a repository, create files or run checks unless the request needs them.",
		"For an explicit request to create or change something, perform the authorized work instead of ending with proposed steps; report only changes and checks actually observed.",
		"Before finishing this same response, check the requested scope and format; distinguish completed work, unverified claims and concrete blockers without adding unrelated work.",
		"Agent assignment is a request to act on the task outcome. Intent labels are heuristic metadata only; do not let missing keywords or task type block execution.",
		"Read the title, current task document and latest user instruction together; start the supported work in this same turn instead of asking the user to repeat an already specified outcome.",
		"Use the user's apparent language and product tone when asking clarification questions or reporting provider limits.",
		"If document content is unavailable, use available task context without inventing requirements; ask only for the specific missing information that prevents useful progress.",
		"When a follow-up thread message supplies a concrete instruction, treat that latest user message as the current directive and use the latest task document context.",
		"Do not ask the user to click local CLI, desktop, or provider approval dialogs.",
		"Do not ask conversational permission questions for file, shell, or tool execution; invoke the tool and let the structured daemon/control-plane approval channel decide.",
		"Ask a conversational question only when a specific missing requirement blocks progress; explain the blocker and ask a focused question, never a generic request for the desired outcome.",
		"If provider usage, quota, or cloud-credit limits stop the work, report the limit as the thread result and do not ask the user for local tool approval.",
	}
}
