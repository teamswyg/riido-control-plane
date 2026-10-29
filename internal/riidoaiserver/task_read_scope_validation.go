package riidoaiserver

func validateTaskReadScope(scope taskReadScope, id, team string) error {
	invalid := taskContextReadError{"invalid"}
	if scope.SchemaVersion != taskReadScopeSchema || scope.RootID != id || scope.TeamID != team || !readScopeStatus(scope.Status) || scope.Selection == nil || scope.Works == nil || scope.Meetings == nil {
		return invalid
	}
	if scope.Selection.RelatedWorkLimit != 5 || scope.Selection.RelatedMeetingLimit != 5 || scope.Selection.ChildrenLimit != 10 {
		return invalid
	}
	seen, counts, root := map[string]bool{}, map[string]int{}, false
	for _, work := range scope.Works {
		if work == nil || work.ID == "" || work.TeamID == "" || work.Kind == AssignmentContentMeeting || !validAssignmentContentKind(work.Kind) || work.Body == nil || work.Comments == nil || work.Links == nil || work.GitHub == nil || work.Repositories == nil || work.PartsStatus == nil || work.Priorities == nil || work.Checklists == nil || work.Attachments == nil || work.Sprints == nil {
			return invalid
		}
		key := string(work.Kind) + "\x00" + work.ID
		if seen[key] {
			return invalid
		}
		seen[key] = true
		if err := readScopeReasons(work.Reasons, work.Kind, counts); err != nil {
			return err
		}
		for _, reason := range work.Reasons {
			if reason.Kind == "root" {
				if work.Kind != AssignmentContentComponent || work.ID != id || work.TeamID != team || root {
					return invalid
				}
				root = true
			}
		}
		if err := validateReadScopeBodySlots(work.ID, work.Body, work.DescriptionBody, work.AlternateBody); err != nil {
			return err
		}
		if err := validateReadScopePartStatus(work.PartsStatus); err != nil {
			return err
		}
		if _, err := readScopeWorkSummary(work); err != nil {
			return err
		}
	}
	for _, meeting := range scope.Meetings {
		if meeting == nil || meeting.ID == "" || meeting.TeamID == "" || seen["meeting\x00"+meeting.ID] {
			return invalid
		}
		seen["meeting\x00"+meeting.ID] = true
		if err := readScopeReasons(meeting.Reasons, AssignmentContentMeeting, counts); err != nil {
			return err
		}
	}
	if !root || counts["related"] > 5 || counts["meeting"] > 5 || counts["child"] > 10 || counts["project"] > 1 || counts["milestone"] > 1 || counts["parentTask"] > 1 {
		return invalid
	}
	return nil
}
