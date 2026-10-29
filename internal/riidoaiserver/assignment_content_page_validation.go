package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
)

func validateAssignmentContentPage(page AssignmentContentPage) error {
	encoded, err := json.Marshal(page)
	data := bytes.TrimSpace(page.Data)
	if err != nil || len(encoded) > assignmentContentMaxPageBytes || !json.Valid(data) || len(data) == 0 || data[0] != '{' || page.Offset < 0 || page.Cursor == "" || page.SourceID == "" || page.Version == "" || !validAssignmentContentPart(page.Part) || !validAssignmentContentKind(page.ResourceKind) {
		return errors.New("assignment content page invalid")
	}
	if (page.Status != "available" && page.Status != "empty" && page.Status != "partial") || len(page.Sources) == 0 {
		return errors.New("assignment content source status invalid")
	}
	for _, source := range page.Sources {
		if source.ID == "" || !assignmentContentAuditKind(source.Kind) || !assignmentContentAuditPart(source.Part) {
			return errors.New("assignment content audit source invalid")
		}
	}
	return nil
}

// These are the existing public AgentContextReadSource enum members.
func assignmentContentAuditKind(kind string) bool {
	switch kind {
	case "task", "project", "milestone", "parent_task", "link", "attachment", "related_task", "related_backlog", "related_intake", "related_meeting", "comment", "target", "agent":
		return true
	}
	return false
}

func assignmentContentAuditPart(part string) bool {
	switch part {
	case "body", "summary", "title", "metadata", "base", "task", "tree", "resources", "comments", "profile":
		return true
	}
	return false
}
