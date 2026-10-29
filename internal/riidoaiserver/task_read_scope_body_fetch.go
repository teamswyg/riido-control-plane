package riidoaiserver

import (
	"context"
	"strings"
)

func (c *AIAgentPrivateTaskContextClient) fetchReadScopeBody(ctx context.Context, bearer string, resource AssignmentScopedResource, first *taskReadBodyPage) (*taskReadBodyPage, error) {
	query, root, idVariable := readScopeBodyQuery, "getComponentReadScopeBody", "componentId"
	switch resource.Kind {
	case AssignmentContentComponent:
		// The component query above is already selected.
	case AssignmentContentBacklog:
		root, idVariable = "getBacklogReadScopeBody", "backlogId"
	case AssignmentContentIntake:
		root, idVariable = "getIntakeReadScopeBody", "intakeId"
	case AssignmentContentMeeting:
		root, idVariable = "getMeetingReadScopeBody", "meetingId"
	}
	// This finite substitution only selects fixed schema roots and ID types.
	query = strings.ReplaceAll(query, "getComponentReadScopeBody", root)
	query = strings.ReplaceAll(query, "componentId", idVariable)
	variables := map[string]any{idVariable: resource.ID, "teamId": resource.TeamID, "part": first.Part, "version": first.Version, "offset": *first.NextOffset, "limit": 8192}
	if first.Selector != nil {
		variables["selector"] = first.Selector
	}
	var page taskReadBodyPage
	if err := c.readScopeGraphQL(ctx, bearer, query, root, variables, &page); err != nil {
		return nil, err
	}
	if err := validateReadScopeBody(&page, resource.ID, first.Part, first.Version, *first.NextOffset); err != nil {
		return nil, err
	}
	if page.Source != first.Source || page.Format != first.Format || readEvidenceKey(page.Selector) != readEvidenceKey(first.Selector) {
		return nil, taskContextReadError{"invalid"}
	}
	return &page, nil
}
