package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (c *AIAgentPrivateTaskContextClient) captureReadScopeWorkMetadata(ctx context.Context, bearer string, b *taskReadScopeBuilder, resource AssignmentScopedResource, work *taskReadScopeWork) (AssignmentContentManifest, error) {
	if work.ContentPage == nil {
		return b.captureWorkMetadata(resource, *work), nil
	}
	raw, complete, err := c.readScopeTypedJSON(ctx, bearer, resource, work.ContentPage, "metadata", taskReadPartSelector{})
	if err != nil {
		return AssignmentContentManifest{}, err
	}
	if !complete {
		return b.captureTypedPrefix(resource, AssignmentContentMetadata, work.ContentPage), nil
	}
	if err := applyReadScopeWorkMetadata(work, raw); err != nil {
		return AssignmentContentManifest{}, err
	}
	return b.captureWorkMetadata(resource, *work), nil
}

func (b *taskReadScopeBuilder) captureTypedPrefix(resource AssignmentScopedResource, part AssignmentContentPart, first *taskReadBodyPage) AssignmentContentManifest {
	return b.addPart(resource, part, first.Version, readScopeTypedPrefixPages(resource, first, 0), false)
}

func readScopeTypedPrefixPages(resource AssignmentScopedResource, first *taskReadBodyPage, index int) []AssignmentContentPage {
	var pages []AssignmentContentPage
	for _, fragment := range readScopeBodyFragments(*first) {
		raw, err := json.Marshal(fragment)
		if err != nil {
			return nil
		}
		pages = append(pages, AssignmentContentPage{Offset: index + len(pages), Status: "partial", Data: raw, Sources: []AssignmentContextSource{readScopeSource(resource, readScopeTypedAuditPart(first.Part), first.Version)}})
	}
	return pages
}

func retainTypedPrefix(first *taskReadBodyPage, raw []byte) {
	offset, next, complete := 0, len(raw), false
	first.Content = string(raw)
	first.Offset = &offset
	first.NextOffset = &next
	first.Complete = &complete
	first.Status = "partial"
}

func readScopeTypedAuditPart(part string) string {
	switch part {
	case "document", "description", "document_alternate":
		return "body"
	case "comments_page":
		return "comments"
	default:
		return "metadata"
	}
}
