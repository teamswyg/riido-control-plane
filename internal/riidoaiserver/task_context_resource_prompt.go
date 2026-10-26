package riidoaiserver

import (
	"strconv"
	"strings"
)

func writeTaskContextResources(out *strings.Builder, resources []AIAgentTaskContextResource, status string, limit int) {
	if len(resources) == 0 && status == "" {
		return
	}
	var section strings.Builder
	header := resourceSectionHeader(status)
	section.WriteString(header)
	lines, omitted := promptResourceLines(resources, limit-len([]rune(header)))
	for _, line := range lines {
		section.WriteString(line.text)
	}
	if omitted > 0 {
		section.WriteString("[" + strconv.Itoa(omitted) + " more link(s) or attachment(s) not shown.]\n")
	}
	// Trimmed like the previous bounded writer so in-budget prompts are unchanged.
	out.WriteString(strings.TrimSpace(strings.ToValidUTF8(section.String(), "\uFFFD")) + "\n")
}

func resourceSectionHeader(status string) string {
	header := "\n### Links and attachments\nMetadata only; contents have not been downloaded or read.\n"
	if status != "available" {
		header += "[Resource metadata incomplete or unavailable.]\n"
	}
	return header
}

// promptResourceLines keeps whole entries within budget runes; the prompt
// and its read records share it so a dropped entry is never recorded as read.
func promptResourceLines(resources []AIAgentTaskContextResource, budget int) ([]promptRelatedLine, int) {
	var lines []promptRelatedLine
	used, omitted := 0, 0
	for _, item := range resources {
		if item.Kind != "file" && item.Kind != "link" {
			continue
		}
		line := "- " + item.Kind + ": " + strings.Join(strings.Fields(boundedTaskContextText(item.Title, 100)), " ")
		if item.ID != "" {
			line += " [id: " + strings.Join(strings.Fields(boundedTaskContextText(item.ID, 100)), " ") + "]"
		}
		if target := safeTaskContextLink(item.URL); target != "" {
			line += " (" + target + ")"
		}
		line += "\n"
		if n := len([]rune(line)); used+n <= budget {
			lines, used = append(lines, promptRelatedLine{text: line, source: assignmentContextResourceSources([]AIAgentTaskContextResource{item})[0]}), used+n
			continue
		}
		omitted++
	}
	return lines, omitted
}

func promptResourceSources(resources []AIAgentTaskContextResource, status string, limit int) []AssignmentContextSource {
	if len(resources) == 0 && status == "" {
		return nil
	}
	lines, _ := promptResourceLines(resources, limit-len([]rune(resourceSectionHeader(status))))
	sources := make([]AssignmentContextSource, 0, len(lines))
	for _, line := range lines {
		sources = append(sources, line.source)
	}
	return sources
}
