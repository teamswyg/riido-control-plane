package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
)

func (c *AIAgentPrivateTaskContextClient) readScopeTypedJSON(ctx context.Context, bearer string, resource AssignmentScopedResource, first *taskReadBodyPage, part string, selector taskReadPartSelector) (json.RawMessage, bool, error) {
	invalid := taskContextReadError{"invalid"}
	if err := validateReadScopeTypedReference(resource, first, part, selector); err != nil {
		return nil, false, err
	}
	expected, _ := normalizeReadScopePartSelector(resource.Kind, part, &selector)
	var err error
	var content strings.Builder
	page := first
	for {
		content.WriteString(page.Content)
		if *page.Complete {
			raw := content.String()
			if !json.Valid([]byte(raw)) || readScopePartVersion(resource.Kind, resource.ID, resource.TeamID, part, first.Source, raw, expected) != first.Version {
				return nil, false, invalid
			}
			return json.RawMessage(raw), true, nil
		}
		if ctx.Err() != nil || content.Len() >= contextCommentItemBytes*assignmentContextMaxCommentPages {
			raw := []byte(content.String())
			retainTypedPrefix(first, raw)
			return raw, false, nil //nolint:nilerr // Preserve an explicitly incomplete capture when the continuation expires or is absent.
		}
		page, err = c.fetchReadScopeBody(ctx, bearer, resource, page)
		if err != nil {
			if readScopeIncomplete(ctx, err) {
				raw := []byte(content.String())
				retainTypedPrefix(first, raw)
				return raw, false, nil
			}
			return nil, false, err
		}
	}
}

func validateReadScopeTypedReference(resource AssignmentScopedResource, first *taskReadBodyPage, part string, selector taskReadPartSelector) error {
	invalid := taskContextReadError{"invalid"}
	expected, err := normalizeReadScopePartSelector(resource.Kind, part, &selector)
	if err != nil || first == nil || first.Source != "riido_"+part || first.Format != "json" || first.Selector == nil {
		return invalid
	}
	actual, err := normalizeReadScopePartSelector(resource.Kind, part, first.Selector)
	if err != nil || readEvidenceKey(expected) != readEvidenceKey(actual) || readEvidenceKey(actual) != readEvidenceKey(first.Selector) {
		return invalid
	}
	return validateReadScopeBody(first, resource.ID, part, "", 0)
}
