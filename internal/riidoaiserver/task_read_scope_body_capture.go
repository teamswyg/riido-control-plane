package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (c *AIAgentPrivateTaskContextClient) captureReadScopeBody(ctx context.Context, bearer string, b *taskReadScopeBuilder, resource AssignmentScopedResource, first *taskReadBodyPage, editor *AIAgentTaskDocumentSnapshot) (AssignmentContentManifest, error) {
	if first == nil || readScopeBodyPart(first.Part) == "" {
		return AssignmentContentManifest{}, taskContextReadError{"invalid"}
	}
	if err := validateReadScopeBody(first, resource.ID, first.Part, "", 0); err != nil {
		return AssignmentContentManifest{}, err
	}
	var pages []AssignmentContentPage
	var content string
	page, complete := first, false
	for {
		content += page.Content
		for _, fragment := range readScopeBodyFragments(*page) {
			data, marshalErr := json.Marshal(fragment)
			if marshalErr != nil {
				return AssignmentContentManifest{}, taskContextReadError{"invalid"}
			}
			pages = append(pages, AssignmentContentPage{Offset: *fragment.Offset, Status: fragment.Status, Data: data, Sources: []AssignmentContextSource{readScopeSource(resource, "body", first.Version)}})
		}
		if *page.Complete {
			complete = true
			break
		}
		if ctx.Err() != nil || len(content) >= contextCommentItemBytes*assignmentContextMaxCommentPages {
			break
		}
		var err error
		if editor != nil && first.Source == "client_snapshot" {
			page, err = readScopeEditorContinuation(resource, *first, editor, *page.NextOffset)
		} else {
			page, err = c.fetchReadScopeBody(ctx, bearer, resource, page)
		}
		if err != nil {
			if readScopeIncomplete(ctx, err) {
				break
			}
			return AssignmentContentManifest{}, err
		}
	}
	if complete && (readScopeBodyVersion(resource.Kind, resource.ID, first.Part, first.Source, first.Format, content) != first.Version || (first.Format == "json" && !json.Valid([]byte(content)))) {
		return AssignmentContentManifest{}, taskContextReadError{"invalid"}
	}
	return b.addPart(resource, readScopeBodyPart(first.Part), first.Version, pages, complete), nil
}
