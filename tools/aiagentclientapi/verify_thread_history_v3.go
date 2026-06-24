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
	if err := requireIdentity(v3.IdentityRules, "conversation_id"); err != nil {
		return err
	}
	if err := requireIdentity(v3.IdentityRules, "thread_id"); err != nil {
		return err
	}
	if err := requireStrings("thread history v3 terminal state", v3.TerminalStates, []string{"completed", "failed", "stopped", "cancelled", "timeout"}); err != nil {
		return err
	}
	return requireStrings("thread history v3 checklist", v3.Checklist, []string{"v3-read-model", "conversation-id-card-key", "v2-mutations", "v2-sse", "terminal-late-event-guard"})
}

func requireIdentity(rules []identityRule, name string) error {
	for _, rule := range rules {
		if rule.Name == name {
			return nil
		}
	}
	return fmt.Errorf("thread history v3 identity %q is required", name)
}
