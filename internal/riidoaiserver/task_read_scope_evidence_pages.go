package riidoaiserver

func (index *taskReadEvidenceIndex) addBody(kind AssignmentContentKind, page *taskReadBodyPage) error {
	if page == nil || page.Offset == nil || index.selected[string(kind)+"\x00"+page.ResourceID] == "" {
		return taskContextReadError{"invalid"}
	}
	if err := validateReadScopeBody(page, page.ResourceID, page.Part, "", *page.Offset); err != nil {
		return err
	}
	key, err := readEvidenceBodyKey(kind, page.ResourceID, page.Part, page.Version, *page.Offset, page.Selector)
	if err != nil {
		return err
	}
	return index.add(key, page)
}

func (index *taskReadEvidenceIndex) addComments(page *taskReadCommentPage) error {
	if page == nil || index.selected[string(page.ResourceKind)+"\x00"+page.ResourceID] == "" {
		return taskContextReadError{"invalid"}
	}
	parent := ""
	if page.ParentCommentID != nil {
		parent = *page.ParentCommentID
	}
	resource := AssignmentScopedResource{Kind: page.ResourceKind, ID: page.ResourceID}
	if err := validateReadScopeComments(page, resource, parent, "", page.Page); err != nil {
		return err
	}
	if page.ContentPage != nil {
		if err := index.addBody(page.ResourceKind, page.ContentPage); err != nil {
			return err
		}
	}
	if err := index.add(readEvidenceKey("comments", page.ResourceKind, page.ResourceID, parent, page.Page, page.Version), page); err != nil {
		return err
	}
	if page.Page == 1 {
		return index.add(readEvidenceKey("comments", page.ResourceKind, page.ResourceID, parent, 1, ""), page)
	}
	return nil
}

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
