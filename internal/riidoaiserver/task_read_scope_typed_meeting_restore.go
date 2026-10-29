package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (c *AIAgentPrivateTaskContextClient) restoreReadScopeMeeting(ctx context.Context, bearer string, ref taskReadScopeMeetingRef, cursor string, out *taskReadMeeting) error {
	resource := AssignmentScopedResource{Kind: AssignmentContentMeeting, ID: ref.ID, TeamID: ref.TeamID}
	if out.ContentPage != nil {
		if out.Summary == nil || out.Summary.Kind != resource.Kind || out.Summary.ID != resource.ID {
			return taskContextReadError{"invalid"}
		}
		raw, complete, err := c.readScopeTypedJSON(ctx, bearer, resource, out.ContentPage, "metadata", taskReadPartSelector{})
		if err != nil {
			return err
		}
		if !complete {
			return errReadScopeCaptureIncomplete
		}
		var metadata taskReadMeetingMetadata
		if decodeReadScopeSlot(raw, &metadata) != nil || metadata.Labels == nil {
			return taskContextReadError{"invalid"}
		}
		var identity struct{ ID string }
		if json.Unmarshal(metadata.Details, &identity) != nil || identity.ID != ref.ID {
			return taskContextReadError{"invalid"}
		}
		out.Details, out.Labels, out.ContentPage = metadata.Details, metadata.Labels, nil
		completeDeferredMetadata(out.PartsStatus)
	}
	if out.Board != nil && out.Board.ContentPage != nil {
		outer := out.Board
		if outer.Complete == nil || *outer.Complete || outer.NextCursor != nil || len(outer.Groups) != 0 || len(outer.Cards) != 0 {
			return taskContextReadError{"invalid"}
		}
		limit := 100
		selector := taskReadPartSelector{Limit: &limit}
		if cursor != "" {
			selector.BoardCursor = &cursor
		}
		raw, complete, err := c.readScopeTypedJSON(ctx, bearer, resource, outer.ContentPage, "board_page", selector)
		if err != nil {
			return err
		}
		if !complete {
			return errReadScopeCaptureIncomplete
		}
		var restored taskReadBoardPage
		if decodeReadScopeSlot(raw, &restored) != nil || restored.ContentPage != nil || restored.Version != outer.Version || restored.Offset == nil || outer.Offset == nil || *restored.Offset != *outer.Offset {
			return taskContextReadError{"invalid"}
		}
		out.Board = &restored
	}
	return nil
}
