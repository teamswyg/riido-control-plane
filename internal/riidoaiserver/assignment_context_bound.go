package riidoaiserver

import (
	"encoding/json"
	"unicode/utf8"
)

// assignmentContextMaxJSONBytes keeps a snapshot inside one DynamoDB item
// (400KB) even if gzip gains nothing and base64 adds a third.
const assignmentContextMaxJSONBytes = 256 << 10

// boundAssignmentContext drops alternate renderings and then halves the
// longest document body until the snapshot fits. It reports whether anything
// was cut so reads are recorded as truncated rather than served.
func boundAssignmentContext(c AIAgentTaskContext) (AIAgentTaskContext, bool) {
	if assignmentContextFits(c) {
		return c, false
	}
	c.Hierarchy = cloneAssignmentContextHierarchy(c.Hierarchy)
	docs := assignmentContextDocuments(&c)
	for _, doc := range docs {
		doc.AlternateContent, doc.AlternateSource, doc.AlternateFormat = "", "", ""
	}
	for range 32 {
		if assignmentContextFits(c) {
			break
		}
		longest := docs[0]
		for _, doc := range docs[1:] {
			if len(doc.Content) > len(longest.Content) {
				longest = doc
			}
		}
		longest.Content = truncateUTF8(longest.Content, len(longest.Content)/2)
	}
	return c, true
}

func assignmentContextFits(c AIAgentTaskContext) bool {
	encoded, err := json.Marshal(c)
	return err == nil && len(encoded) <= assignmentContextMaxJSONBytes
}

func assignmentContextDocuments(c *AIAgentTaskContext) []*AIAgentTaskContextDocument {
	docs := []*AIAgentTaskContextDocument{&c.Document}
	for _, ref := range []*AIAgentTaskContextReference{&c.Hierarchy.Project, &c.Hierarchy.Milestone, &c.Hierarchy.ParentTask} {
		if ref.Document != nil {
			docs = append(docs, ref.Document)
		}
	}
	return docs
}

// The caller's context must not be mutated through shared document pointers.
func cloneAssignmentContextHierarchy(h AIAgentTaskContextHierarchy) AIAgentTaskContextHierarchy {
	for _, ref := range []*AIAgentTaskContextReference{&h.Project, &h.Milestone, &h.ParentTask} {
		if ref.Document != nil {
			doc := *ref.Document
			ref.Document = &doc
		}
	}
	return h
}

func truncateUTF8(s string, n int) string {
	if n >= len(s) {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}
