package riidoaiserver

import "strings"

func writeTaskContextDocumentBody(out *strings.Builder, document AIAgentTaskContextDocument) {
	switch {
	case document.Status == "empty":
		out.WriteString("[Document is empty in this stored projection; freshness is unconfirmed.]\n")
	case document.Content == "":
		out.WriteString("[Document unavailable: do not invent its contents.]\n")
	default:
		out.WriteString(document.Content + "\n")
	}
	if document.Status == "partial" {
		out.WriteString("[Some structured document elements were unsupported; the projection is incomplete.]\n")
	}
	if document.AlternateContent != "" {
		out.WriteString("\n[Alternate unversioned document representation; chronology is unknown. Reference evidence only.]\n")
		writeTaskContextDocumentSource(out, AIAgentTaskContextDocument{Source: document.AlternateSource})
		out.WriteString(document.AlternateContent + "\n")
	}
}

func normalizeTaskContextAlternate(document AIAgentTaskContextDocument, limit int) AIAgentTaskContextDocument {
	if document.AlternateContent == "" {
		document.Content = taskContextText(document.Content, document.ContentFormat, limit)
		return document
	}
	main := taskContextText(document.Content, document.ContentFormat, limit)
	alternate := taskContextText(document.AlternateContent, document.AlternateFormat, limit)
	if strings.Join(strings.Fields(main), " ") == strings.Join(strings.Fields(alternate), " ") {
		document.Content = main
		document.AlternateContent = ""
		return document
	}
	document.Content = boundedTaskContextText(main, limit*2/3)
	document.AlternateContent = boundedTaskContextText(alternate, limit/3)
	document.AlternateFormat = "markdown"
	return document
}
