package riidoaiserver

import (
	"context"
)

func (c *AIAgentPrivateTaskContextClient) restoreReadScopeComments(ctx context.Context, bearer string, resource AssignmentScopedResource, outer *taskReadCommentPage) (*taskReadCommentPage, error) {
	if outer == nil || outer.ContentPage == nil {
		return outer, nil
	}
	parent := ""
	if outer.ParentCommentID != nil {
		parent = *outer.ParentCommentID
	}
	selector := taskReadPartSelector{ParentCommentID: outer.ParentCommentID, Page: &outer.Page, Limit: &outer.Limit}
	if err := validateReadScopeComments(outer, resource, parent, outer.Version, outer.Page); err != nil {
		return nil, err
	}
	raw, complete, err := c.readScopeTypedJSON(ctx, bearer, resource, outer.ContentPage, "comments_page", selector)
	if err != nil {
		return nil, err
	}
	if !complete {
		return nil, errReadScopeCaptureIncomplete
	}
	var restored taskReadCommentPage
	if decodeReadScopeSlot(raw, &restored) != nil || restored.ContentPage != nil || restored.Limit != outer.Limit {
		return nil, taskContextReadError{"invalid"}
	}
	if err := validateReadScopeComments(&restored, resource, parent, outer.Version, outer.Page); err != nil {
		return nil, err
	}
	return &restored, nil
}
