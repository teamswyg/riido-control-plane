package riidoaiserver

import "strings"

func writeTaskContextDocumentSource(out *strings.Builder, document AIAgentTaskContextDocument) {
	switch document.Source {
	case "provider_document", "provider_json_projection", "provider_html_projection":
		writePromptLine(out, "content_source", document.Source)
		out.WriteString("[Stored provider projection; freshness and current Yjs parity are unconfirmed.]\n")
	case "component_snapshot":
		writePromptLine(out, "content_source", "component_snapshot")
		out.WriteString("[Component snapshot freshness is unconfirmed.]\n")
	}
}
