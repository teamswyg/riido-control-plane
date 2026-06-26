package main

import "fmt"

func verifyThreadHistoryV3(v3 threadHistoryV3) error {
	if v3.ReadEndpoint.Path != "/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads" {
		return fmt.Errorf("thread history v3 read endpoint is required")
	}
	if v3.SSEEndpoint.Path != "/v2/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/thread-stream-subscription" {
		return fmt.Errorf("thread history v3 must keep v2 SSE subscription")
	}
	if len(v3.ActionEndpoints) < 4 {
		return fmt.Errorf("thread history v3 action endpoints are incomplete")
	}
	if err := requireNamedRules("thread history v3 implementation step", v3.ImplementationSteps, []string{"initial-load-v3", "group-by-conversation-id", "optimistic-v2-mutation", "refetch-v3-after-mutation"}); err != nil {
		return err
	}
	if err := requireShape(v3.ResponseShapes, "AIAgentTaskThreadHistoryRecord", []string{"conversation_id", "parent_thread_id", "messages[]"}); err != nil {
		return err
	}
	if err := requireShape(v3.ResponseShapes, "AIAgentTaskThreadHistoryMessage", []string{"message_id", "role", "body", "result_message"}); err != nil {
		return err
	}
	if err := requireIdentity(v3.IdentityRules, "conversation_id"); err != nil {
		return err
	}
	if err := requireIdentity(v3.IdentityRules, "thread_id"); err != nil {
		return err
	}
	if err := requireNamedRules("thread history v3 ordering rule", v3.OrderingRules, []string{"conversation-card-ordering", "message-row-ordering", "queued-status-current-only", "late-terminal-guard"}); err != nil {
		return err
	}
	if err := requireNamedRules("thread history v3 mutation rule", v3.MutationRules, []string{"assign-agent", "send-thread-message", "stop-agent", "delete-agent"}); err != nil {
		return err
	}
	scenarios := []string{
		"intent-clarification-before-deliverable",
		"intent-clarification-waits-for-user",
		"concrete-followup-authoritative",
		"draft-then-research-limit-in-same-conversation",
		"provider-limit-result",
	}
	if err := requireNamedRules("thread history v3 interaction scenario", v3.InteractionScenarios, scenarios); err != nil {
		return err
	}
	if err := requireStrings("thread history v3 terminal state", v3.TerminalStates, []string{"completed", "failed", "stopped", "cancelled", "timeout"}); err != nil {
		return err
	}
	return requireStrings("thread history v3 checklist", v3.Checklist, []string{"v3-read-model", "conversation-id-card-key", "v2-mutations", "v2-sse", "queued-status-current-only", "terminal-late-event-guard"})
}

func requireIdentity(rules []identityRule, name string) error {
	for _, rule := range rules {
		if rule.Name == name {
			return nil
		}
	}
	return fmt.Errorf("thread history v3 identity %q is required", name)
}
