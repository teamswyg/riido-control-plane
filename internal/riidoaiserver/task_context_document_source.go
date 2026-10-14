package riidoaiserver

import "strings"

func writeTaskContextDocumentSource(out *strings.Builder, document AIAgentTaskContextDocument) {
	switch document.Source {
	case "provider_document":
		writePromptLine(out, "content_source", "provider_document")
	case "component_snapshot":
		writePromptLine(out, "content_source", "component_snapshot")
		out.WriteString("[Fresh provider document unavailable; component snapshot freshness is unconfirmed.]\n")
	}
}
