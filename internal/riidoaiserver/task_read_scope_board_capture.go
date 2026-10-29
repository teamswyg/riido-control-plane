package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (c *AIAgentPrivateTaskContextClient) captureReadScopeBoard(ctx context.Context, bearer string, b *taskReadScopeBuilder, resource AssignmentScopedResource, ref taskReadScopeMeetingRef, first *taskReadMeeting) (AssignmentContentManifest, error) {
	page := first
	var pages []AssignmentContentPage
	bytes := 0
	complete := false
	seen := map[string]bool{}
	for {
		for _, card := range page.Board.Cards {
			var identity struct {
				ID string `json:"id"`
			}
			if json.Unmarshal(card, &identity) != nil || identity.ID == "" || seen[identity.ID] {
				return AssignmentContentManifest{}, taskContextReadError{"invalid"}
			}
			seen[identity.ID] = true
		}
		data, marshalErr := json.Marshal(page.Board)
		if marshalErr != nil {
			return AssignmentContentManifest{}, taskContextReadError{"invalid"}
		}
		bytes += len(data)
		pages = append(pages, readScopeJSONPages(data, *page.Board.Offset, len(pages), []AssignmentContextSource{readScopeSource(resource, "metadata", page.Board.Version)})...)
		if *page.Board.Complete {
			complete = true
			break
		}
		if ctx.Err() != nil || bytes >= contextCommentItemBytes*assignmentContextMaxCommentPages {
			break
		}
		next, err := c.fetchReadScopeMeeting(ctx, bearer, ref, *page.Board.NextCursor, page)
		if err != nil {
			if readScopeIncomplete(ctx, err) {
				if next != nil && next.Board != nil && next.Board.ContentPage != nil {
					pages = append(pages, readScopeTypedPrefixPages(resource, next.Board.ContentPage, len(pages))...)
				}
				break
			}
			return AssignmentContentManifest{}, err
		}
		page = next
	}
	return b.addPart(resource, AssignmentContentBoard, first.Board.Version, pages, complete), nil //nolint:nilerr // A deadline preserves an explicitly incomplete captured prefix.
}
