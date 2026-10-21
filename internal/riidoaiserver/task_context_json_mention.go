package riidoaiserver

func taskContextRiidoMentionLabel(node taskContextJSONNode) string {
	for _, name := range []string{"label", "mentionId", "componentId", "id"} {
		if value := taskContextJSONAttribute(node.Attrs, name); value != "" {
			return "@" + value
		}
	}
	return "[Mention]"
}

func taskContextStandardMentionLabel(node taskContextJSONNode) string {
	for _, name := range []string{"label", "name", "id", "mentionId", "componentId"} {
		if value := taskContextJSONAttribute(node.Attrs, name); value != "" {
			return value
		}
	}
	return "Mention"
}
