package riidoaiserver

func assignmentContextTestTaskContext() AIAgentTaskContext {
	return AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{ID: "task-private", ComponentType: "task", Title: "Riido CLI 개발", KeyNumber: "RIID-5900"},
		Document:  AIAgentTaskContextDocument{Content: "<p>task body</p>", ContentFormat: "html"},
		Resources: []AIAgentTaskContextResource{{ID: "link-1", Kind: "link", Title: "spec", URL: "https://example.com/spec"}},
		Hierarchy: AIAgentTaskContextHierarchy{
			Project:   AIAgentTaskContextReference{ID: "project-a", Title: "AI Agent", KeyNumber: "RIID-4800", Document: &AIAgentTaskContextDocument{Content: "<p>project body</p>"}},
			Milestone: AIAgentTaskContextReference{ID: "milestone-a", Title: "CLI", KeyNumber: "RIID-5800"},
		},
	}
}
