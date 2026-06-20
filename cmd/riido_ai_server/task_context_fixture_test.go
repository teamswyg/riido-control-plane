package main

import "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

func openAPITaskContextFixture() riidoaiserver.AIAgentTaskContext {
	return riidoaiserver.AIAgentTaskContext{
		Component: riidoaiserver.AIAgentTaskContextComponent{
			ID:         "component-a",
			Title:      "Task context from existing API server",
			BranchName: "RIID-4800-server-task-context-http-client-assignment-prompt-wiring",
		},
		Document: riidoaiserver.AIAgentTaskContextDocument{
			Content:       "Existing API server document markdown.",
			ContentFormat: "markdown",
		},
		Hierarchy:    riidoaiserver.AIAgentTaskContextHierarchy{},
		Repositories: []riidoaiserver.AIAgentTaskContextRepository{},
	}
}
