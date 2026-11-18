package riidoaiserver

func (index *taskReadEvidenceIndex) addMeeting(captured taskReadCapturedMeeting) error {
	meeting := captured.Snapshot
	if meeting == nil || meeting.TeamID == "" || meeting.TeamID != index.selected["meeting\x00"+meeting.ID] {
		return taskContextReadError{"invalid"}
	}
	cursor := ""
	if captured.RequestCursor != nil {
		cursor = *captured.RequestCursor

		if cursor == "" {
			return taskContextReadError{"invalid"}
		}
	}
	if meeting.Board != nil && meeting.Board.ContentPage != nil {
		if err := index.addBody(AssignmentContentMeeting, meeting.Board.ContentPage); err != nil {
			return err
		}
	}
	if err := index.add(readEvidenceKey("meeting", meeting.ID, cursor), meeting); err != nil {
		return err
	}
	if cursor == "" {
		for _, page := range []*taskReadBodyPage{meeting.Body, meeting.DescriptionBody, meeting.AlternateBody, meeting.ContentPage} {
			if page != nil {
				if err := index.addBody(AssignmentContentMeeting, page); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
