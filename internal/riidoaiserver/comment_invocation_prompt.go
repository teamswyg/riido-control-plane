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
	prompt.WriteString("\n\nBefore selecting another Riido Agent, run riido agents list --json --no-input and riido agents show <agent-id> --json --no-input for the candidate. Follow next_cursor to inspect further candidates when needed. Compare public description, role_status, work_status, runtime.connection and callable/unavailable_reasons against the work needed. Never infer expertise from a name or provider. If the CLI is unavailable or role evidence is insufficient, ask the user who should perform that work; do not guess a delegation. The syntax catalog below is an identity hint, not current role/availability evidence. The server rechecks access and execution availability.\n")
	prompt.WriteString("\n\nRiido comment handoff: If another Agent must perform work, return an HTML final response containing its typed mention from the catalog below plus an explicit request. Plain @names, quoted/code examples, progress and imported history do not execute. Do not mention an Agent merely to describe it. Use only catalog IDs. Self/ancestor calls are rejected; at most 3 automatic hops and 8 additional runs per initiating request.\n")
	for _, agent := range agents.Agents {
		if agent.AgentID == req.AgentID {
			continue
		}
		fmt.Fprintf(&prompt, "<span data-type=\"agent-mention\" data-mention-kind=\"agent\" data-mention-id=\"%s\">%s</span>\n", html.EscapeString(agent.AgentID), html.EscapeString(agent.Name))
	}
	req.Prompt += prompt.String()
	return nil
}
