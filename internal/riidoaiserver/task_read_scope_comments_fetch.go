package riidoaiserver

import "context"

func (c *AIAgentPrivateTaskContextClient) fetchReadScopeComments(ctx context.Context, bearer string, resource AssignmentScopedResource, parent, version string, page, limit int) (*taskReadCommentPage, error) {
	variables := map[string]any{"resourceId": resource.ID, "resourceKind": string(resource.Kind), "teamId": resource.TeamID, "page": page, "limit": limit, "parentCommentId": nil, "version": nil}
	if parent != "" {
		variables["parentCommentId"] = parent
	}
	if version != "" {
		variables["version"] = version
	}
	var out taskReadCommentPage
	if err := c.readScopeGraphQL(ctx, bearer, readScopeCommentsQuery, "getReadScopeComments", variables, &out); err != nil {
		return nil, err
	}
	restored, err := c.restoreReadScopeComments(ctx, bearer, resource, &out)
	if err != nil {
		return &out, err
	}
	out = *restored
	if err := validateReadScopeComments(&out, resource, parent, version, page); err != nil {
		return nil, err
	}
	if out.Limit != limit {
		return nil, taskContextReadError{"invalid"}
	}
	return &out, nil
}

func validateReadScopeComments(page *taskReadCommentPage, resource AssignmentScopedResource, parent, version string, number int) error {
	if page == nil || page.ResourceID != resource.ID || page.ResourceKind != resource.Kind || page.Page != number || page.Limit < 1 || page.Limit > 100 || page.Version == "" || (version != "" && page.Version != version) || page.Items == nil || page.Complete == nil || len(page.Items) > page.Limit {
		return taskContextReadError{"invalid"}
	}
	actualParent := ""
	if page.ParentCommentID != nil {
		actualParent = *page.ParentCommentID
	}
	if actualParent != parent {
		return taskContextReadError{"invalid"}
	}
	if page.ContentPage != nil {
		if *page.Complete || page.NextPage != nil || len(page.Items) != 0 || page.Status != "deferred" {
			return taskContextReadError{"invalid"}
		}
		return validateReadScopeTypedReference(resource, page.ContentPage, "comments_page", taskReadPartSelector{ParentCommentID: page.ParentCommentID, Page: &page.Page, Limit: &page.Limit})
	}
	if *page.Complete {
		if page.NextPage != nil || (page.Status != "available" && page.Status != "empty") {
			return taskContextReadError{"invalid"}
		}
	} else if page.NextPage == nil || *page.NextPage != number+1 || len(page.Items) != page.Limit || page.Status != "partial" {
		return taskContextReadError{"invalid"}
	}
	return nil
}
