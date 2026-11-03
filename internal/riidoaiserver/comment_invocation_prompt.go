package riidoaiserver

import (
	"context"
	"fmt"
	"html"
	"strings"
)

func (s Server) appendCommentInvocationPrompt(ctx context.Context, principal AuthorizationResult, taskID string, req *AssignRequest) error {
	agents, err := s.aiAgent.ListAIAgentTaskAssignableAgents(ctx, principal, taskID)
	if err != nil {
		return err
	}
	var prompt strings.Builder
	prompt.WriteString(`

## Decide who should do the work
Use the current request, your assigned role and instructions, and what has already been done in this conversation. Answer directly when you can complete the work within your role. A routine edit, summary, or follow-up usually needs no handoff. Respect the requester's scope, including a request to work alone.
Take the initiative when the outcome depends on a different specialty, a named responsibility in your own instructions, or an independent check that you cannot supply yourself. For example, a content writer can revise a headline directly; if a draft depends on an unvalidated problem definition, a problem-definition specialist can review that premise while the writer supplies the draft context. Apply this judgment to the actual roles; do not assume these roles exist.
When there is such a concrete need, check the live directory even if the user did not name an Agent. Run riido agents list --json --no-input, follow next_cursor as needed, and inspect the candidate with riido agents show <agent-id> --json --no-input. Match its public description and role_status to the specific work; check work_status, runtime.connection and unavailable_reasons. Delegate only when policy_checked and callable are both true. Never infer a role from a name or provider. When one eligible Agent clearly fits, invoke it now without asking the user to select or approve it. Do not hand off work already done or send a request back to the caller. If no verified match is available, do the part you can and ask a focused question only if the remaining work is blocked.

## What to put in the Riido conversation
Keep directory checks internal, including in user-visible progress. In the final HTML reply, use the selected Agent's typed mention from the catalog once, say briefly why its role is needed for this task, and ask for a specific result with any context it needs. Use the conversation's language. Leave out commands, tool names, raw IDs, candidate tables and status flags unless the user explicitly asks for diagnostics. Do not claim the other Agent has done the work before it replies. When answering directly, just give the result.
Use a typed mention only to actually invoke an Agent. Plain @names, examples, progress and imported history do not execute. The catalog below supplies mention syntax only. The server rechecks permission and availability and rejects self/ancestor calls; limits are 3 automatic hops and 8 additional runs per initiating request.
`)
	for _, agent := range agents.Agents {
		if agent.AgentID == req.AgentID {
			continue
		}
		fmt.Fprintf(&prompt, "<span data-type=\"agent-mention\" data-mention-kind=\"agent\" data-mention-id=\"%s\">%s</span>\n", html.EscapeString(agent.AgentID), html.EscapeString(agent.Name))
	}
	req.Prompt += prompt.String()
	return nil
}
