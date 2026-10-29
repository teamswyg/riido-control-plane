package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (c *AIAgentPrivateTaskContextClient) captureReadScopeComments(ctx context.Context, bearer string, b *taskReadScopeBuilder, resource AssignmentScopedResource, first *taskReadCommentPage) (AssignmentContentManifest, error) {
	restored, restoreErr := c.restoreReadScopeComments(ctx, bearer, resource, first)
	if restoreErr != nil {
		if readScopeIncomplete(ctx, restoreErr) && first != nil && first.ContentPage != nil {
			return b.captureTypedPrefix(resource, AssignmentContentComments, first.ContentPage), nil
		}
		return AssignmentContentManifest{}, restoreErr
	}
	first = restored
	if err := validateReadScopeComments(first, resource, "", "", 1); err != nil {
		return AssignmentContentManifest{}, err
	}
	page := first
	pages := []AssignmentContentPage{}
	parents := []string{}
	seen := map[string]bool{}
	complete := false
	bytes := 0
	for {
		sources, replies, err := readScopeCommentSources(page, resource)
		if err != nil {
			return AssignmentContentManifest{}, err
		}
		for _, parent := range replies {
			if !seen[parent] {
				seen[parent] = true
				parents = append(parents, parent)
			}
		}
		data, marshalErr := json.Marshal(page)
		if marshalErr != nil {
			return AssignmentContentManifest{}, taskContextReadError{"invalid"}
		}
		bytes += len(data)
		pages = append(pages, readScopeJSONPages(data, page.Page, len(pages), sources)...)
		if ctx.Err() != nil || bytes >= contextCommentItemBytes*assignmentContextMaxCommentPages {
			break
		}
		parent, version, next := "", page.Version, 0
		if page.ParentCommentID != nil {
			parent = *page.ParentCommentID
		}
		switch {
		case !*page.Complete:
			next = *page.NextPage
		case len(parents) > 0:
			parent, parents = parents[0], parents[1:]
			version = ""
			next = 1
		default:
			complete = true
		}
		if complete {
			break
		}
		page, err = c.fetchReadScopeComments(ctx, bearer, resource, parent, version, next, first.Limit)
		if err != nil {
			if readScopeIncomplete(ctx, err) {
				if page != nil && page.ContentPage != nil {
					pages = append(pages, readScopeTypedPrefixPages(resource, page.ContentPage, len(pages))...)
				}
				break
			}
			return AssignmentContentManifest{}, err
		}
	}
	return b.addPart(resource, AssignmentContentComments, first.Version, pages, complete), nil //nolint:nilerr // A deadline preserves an explicitly incomplete captured prefix.
}
