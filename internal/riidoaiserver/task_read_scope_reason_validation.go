package riidoaiserver

func readScopeReasons(reasons []taskReadScopeReason, kind AssignmentContentKind, counts map[string]int) error {
	if len(reasons) == 0 {
		return taskContextReadError{"invalid"}
	}
	seen := map[string]bool{}
	counted := map[string]bool{}
	for _, reason := range reasons {
		key := reason.Kind
		if reason.Source != nil {
			if reason.Kind != "mention" || (*reason.Source != "document" && *reason.Source != "document_alternate") {
				return taskContextReadError{"invalid"}
			}
			key += "\x00" + *reason.Source
		}
		if reason.Relation != nil {
			key += "\x00" + *reason.Relation
		}
		if seen[key] {
			return taskContextReadError{"invalid"}
		}
		seen[key] = true
		if kind == AssignmentContentMeeting && reason.Kind != "meeting" {
			return taskContextReadError{"invalid"}
		}
		switch reason.Kind {
		case "root", "parentTask", "milestone", "project", "child":
			if kind != AssignmentContentComponent || reason.Relation != nil {
				return taskContextReadError{"invalid"}
			}
		case "mention", "client_snapshot_mention", "root_comment_mention", "trigger_comment_mention":
			if kind == AssignmentContentMeeting || reason.Relation != nil {
				return taskContextReadError{"invalid"}
			}
		case "meeting":
			if kind != AssignmentContentMeeting || reason.Relation != nil {
				return taskContextReadError{"invalid"}
			}
		case "related":
			if reason.Relation == nil || !readScopeRelation(*reason.Relation) {
				return taskContextReadError{"invalid"}
			}
		default:
			return taskContextReadError{"invalid"}
		}
		if !counted[reason.Kind] {
			counts[reason.Kind]++
			counted[reason.Kind] = true
		}
	}
	return nil
}

func readScopeRelation(relation string) bool {
	switch relation {
	case "blocking", "blockedBy", "hasDuplicate", "duplicatedOf", "relatedTo":
		return true
	}
	return false
}
