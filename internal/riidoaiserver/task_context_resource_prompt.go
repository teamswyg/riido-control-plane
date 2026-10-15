package riidoaiserver

import "strings"

func writeTaskContextResources(out *strings.Builder, resources []AIAgentTaskContextResource, status string, limit int) {
	if len(resources) == 0 && status == "" {
		return
	}
	var section strings.Builder
	section.WriteString("\n### Links and attachments\nMetadata only; contents have not been downloaded or read.\n")
	if status != "available" {
		section.WriteString("[Resource metadata incomplete or unavailable.]\n")
	}
	for _, item := range resources {
		if item.Kind != "file" && item.Kind != "link" {
			continue
		}
		section.WriteString("- " + item.Kind + ": " + strings.Join(strings.Fields(boundedTaskContextText(item.Title, 100)), " "))
		if item.ID != "" {
			section.WriteString(" [id: " + strings.Join(strings.Fields(boundedTaskContextText(item.ID, 100)), " ") + "]")
		}
		if target := safeTaskContextLink(item.URL); target != "" {
			section.WriteString(" (" + target + ")")
		}
		section.WriteString("\n")
	}
	out.WriteString(boundedTaskContextText(section.String(), limit) + "\n")
}
