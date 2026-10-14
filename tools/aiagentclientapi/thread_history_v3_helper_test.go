package main

func validThreadHistoryV3() threadHistoryV3 {
	return threadHistoryV3{
		ReadEndpoint:    endpointRule{Path: "/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads"},
		SSEEndpoint:     endpointRule{Path: "/v2/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/thread-stream-subscription"},
		ActionEndpoints: []endpointRule{{}, {}, {}, {}},
		ImplementationSteps: namedRules("initial-load-v3", "group-by-conversation-id",
			"optimistic-v2-mutation", "refetch-v3-after-mutation"),
		ResponseShapes: []shapeRule{
			{Name: "AIAgentTaskThreadHistoryRecord", Fields: []string{"conversation_id", "parent_thread_id", "messages[]"}},
			{Name: "AIAgentTaskThreadHistoryMessage", Fields: []string{"message_id", "role", "author_principal_id", "body", "result_message"}},
		},
		IdentityRules: []identityRule{{Name: "conversation_id"}, {Name: "thread_id"}},
		OrderingRules: namedRules("conversation-card-ordering", "message-row-ordering",
			"queued-status-current-only", "queued-stream-current-only",
			"late-terminal-guard", "terminal-active-stream-closure"),
		MutationRules: namedRules("assign-agent", "send-thread-message", "stop-agent", "delete-agent"),
		InteractionScenarios: namedRules("intent-clarification-before-deliverable",
			"intent-clarification-waits-for-user", "concrete-followup-authoritative",
			"draft-then-research-limit-in-same-conversation", "provider-limit-result"),
		TerminalStates: []string{"completed", "failed", "stopped", "cancelled", "timeout"},
		Checklist: []string{
			"v3-read-model", "conversation-id-card-key", "v2-mutations",
			"v2-sse", "queued-status-current-only", "queued-stream-current-only",
			"terminal-late-event-guard", "terminal-active-stream-closure",
		},
	}
}

func namedRules(names ...string) []namedRule {
	rules := make([]namedRule, 0, len(names))
	for _, name := range names {
		rules = append(rules, namedRule{Name: name, Detail: "required"})
	}
	return rules
}
