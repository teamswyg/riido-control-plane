package riidoaiserver

func assignmentContextTestTaskContext() AIAgentTaskContext {
	return AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{ID: "task-private", ComponentType: "task", Title: "Riido CLI 개발", KeyNumber: "RIID-5900", UpdatedAt: "2026-10-25T09:00:00Z"},
		Document:  AIAgentTaskContextDocument{Content: "<p>task body</p>", ContentFormat: "html"},
		Resources: []AIAgentTaskContextResource{{ID: "link-1", Kind: "link", Title: "spec", URL: "https://example.com/spec"}},
		Hierarchy: AIAgentTaskContextHierarchy{
			Project:   AIAgentTaskContextReference{ID: "project-a", Title: "AI Agent", KeyNumber: "RIID-4800", UpdatedAt: "2026-10-01T00:00:00+09:00", Document: &AIAgentTaskContextDocument{Content: "<p>project body</p>"}},
			Milestone: AIAgentTaskContextReference{ID: "milestone-a", Title: "CLI", KeyNumber: "RIID-5800"},
		},
	}
}
