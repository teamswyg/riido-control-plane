package riidoaiserver

import "strings"

type privateDocumentProjections struct {
	HTML *string `json:"html"`
	JSON *string `json:"tiptapJson"`
}

func (p privateDocumentProjections) document() *AIAgentTaskContextDocument {
	if p.JSON != nil && strings.TrimSpace(*p.JSON) != "" {
		if text, complete, err := readableTaskContextJSON(*p.JSON); err == nil {
			return projectionContextDocument(text, "provider_json_projection", complete)
		}
	}
	if p.HTML != nil {
		return projectionContextDocument(readableTaskContextHTML(*p.HTML), "provider_html_projection", true)
	}
	return nil
}

func projectionContextDocument(text, source string, complete bool) *AIAgentTaskContextDocument {
	status := "available"
	if strings.TrimSpace(text) == "" {
		status = "empty"
	}
	if !complete {
		status = "partial"
	}
	return &AIAgentTaskContextDocument{Content: text, ContentFormat: "markdown", Source: source, Status: status}
}

func selectTaskContextDocument(base AIAgentTaskContextDocument, projection *AIAgentTaskContextDocument) AIAgentTaskContextDocument {
	base.Source = "component_snapshot"
	base.Status = "unavailable"
	if strings.TrimSpace(base.Content) != "" {
		base.Status = "available"
	}
	if projection == nil {
		return base
	}
	candidate := *projection
	candidate.ID, candidate.TiptapDocumentID = base.ID, base.TiptapDocumentID
	if candidate.Status == "empty" {
		return candidate
	}
	if base.Status == "available" && strings.Join(strings.Fields(taskContextText(base.Content, base.ContentFormat, taskDocumentContextLimit)), " ") != strings.Join(strings.Fields(candidate.Content), " ") {
		if candidate.Source == "provider_html_projection" {
			base.AlternateContent, base.AlternateFormat, base.AlternateSource = candidate.Content, candidate.ContentFormat, candidate.Source
			return base
		}
		candidate.AlternateContent, candidate.AlternateFormat, candidate.AlternateSource = base.Content, base.ContentFormat, base.Source
	}
	return candidate
}
