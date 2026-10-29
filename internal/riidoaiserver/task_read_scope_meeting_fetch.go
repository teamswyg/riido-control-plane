package riidoaiserver

import (
	"bytes"
	"context"
)

func (c *AIAgentPrivateTaskContextClient) fetchReadScopeMeeting(ctx context.Context, bearer string, ref taskReadScopeMeetingRef, cursor string, previous *taskReadMeeting) (*taskReadMeeting, error) {
	variables := map[string]any{"meetingId": ref.ID, "teamId": ref.TeamID, "boardCursor": nil, "boardLimit": 100}
	if cursor != "" {
		variables["boardCursor"] = cursor
	}
	var out taskReadMeeting
	if err := c.readScopeGraphQL(ctx, bearer, meetingScopeQuery, "getMeetingReadScope", variables, &out); err != nil {
		return nil, err
	}
	if out.ID != ref.ID || out.TeamID != ref.TeamID || !readScopeStatus(out.Status) || out.Board == nil || out.Body == nil || out.DescriptionBody == nil || out.Labels == nil || out.PartsStatus == nil {
		return nil, taskContextReadError{"invalid"}
	}
	if err := validateReadScopePartStatus(out.PartsStatus); err != nil {
		return nil, err
	}
	if err := validateReadScopeBodySlots(out.ID, out.Body, out.DescriptionBody, out.AlternateBody); err != nil {
		return nil, err
	}
	if err := c.restoreReadScopeMeeting(ctx, bearer, ref, cursor, &out); err != nil {
		return &out, err
	}
	board := out.Board
	if board.Version == "" || board.Offset == nil || board.Complete == nil || board.Groups == nil || board.Cards == nil || len(board.Cards) > 100 {
		return nil, taskContextReadError{"invalid"}
	}
	if *board.Complete {
		if board.NextCursor != nil {
			return nil, taskContextReadError{"invalid"}
		}
	} else if board.NextCursor == nil || *board.NextCursor == "" || *board.NextCursor == cursor || len(board.Cards) == 0 {
		return nil, taskContextReadError{"invalid"}
	}
	expected := 0
	if previous != nil {
		expected = *previous.Board.Offset + len(previous.Board.Cards)
		currentMetadata, currentErr := readScopeMeetingStableMetadata(out)
		previousMetadata, previousErr := readScopeMeetingStableMetadata(*previous)
		if currentErr != nil || previousErr != nil || board.Version != previous.Board.Version || !bytes.Equal(currentMetadata, previousMetadata) {
			return nil, taskContextReadError{"invalid"}
		}
	}
	if *board.Offset != expected {
		return nil, taskContextReadError{"invalid"}
	}
	return &out, nil
}

func readScopeMeetingStableMetadata(meeting taskReadMeeting) ([]byte, error) {
	meeting.Status = ""       // Aggregate status changes when the last board page becomes complete.
	meeting.PartsStatus = nil // page completion changes while the underlying metadata remains fixed.
	return readScopeMeetingMetadata(meeting)
}
