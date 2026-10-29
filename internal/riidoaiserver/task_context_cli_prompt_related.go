package riidoaiserver

import "strings"

const taskContextRelatedPromptRunes = 2500

type promptRelatedLine struct {
	text   string
	source AssignmentContextSource
}

// promptRelatedLines renders whole related-item lines within the prompt
// budget. The prompt and its read records both use this list, so an item
// left out of the prompt is never recorded as read.
func promptRelatedLines(extended *AIAgentTaskContextExtended) ([]promptRelatedLine, int) {
	var lines []promptRelatedLine
	used, omitted := 0, 0
	add := func(text string, source AssignmentContextSource) {
		if n := len([]rune(text)); used+n <= taskContextRelatedPromptRunes {
			lines, used = append(lines, promptRelatedLine{text: text, source: source}), used+n
			return
		}
		omitted++
	}
	if extended == nil {
		return nil, 0
	}
	for _, work := range extended.RelatedWork {
		add("- "+work.Kind+" ("+work.Relation+"): "+strings.Join(strings.Fields(work.Title), " ")+" [id: "+work.ID+"]\n",
			AssignmentContextSource{Kind: "related_" + strings.TrimSpace(work.Kind), ID: work.ID, Part: "title"})
	}
	for _, meeting := range extended.RelatedMeetings {
		add("- meeting: "+strings.Join(strings.Fields(meeting.Title), " ")+" [id: "+meeting.ID+"]\n",
			AssignmentContextSource{Kind: "related_meeting", ID: meeting.ID, Part: "title"})
	}
	return lines, omitted
}
