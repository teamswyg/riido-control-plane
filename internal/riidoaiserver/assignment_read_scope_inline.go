package riidoaiserver

import (
	"context"
	"encoding/json"
)

const assignmentScopedInlineBytes = 48 << 10

type assignmentScopedInlinePart struct {
	Manifest   AssignmentContentManifest `json:"manifest"`
	Page       *AssignmentContentPage    `json:"page,omitempty"`
	NextCursor string                    `json:"next_cursor,omitempty"`
}

type assignmentScopedInlineView struct {
	Resource AssignmentScopedResource     `json:"resource"`
	Parts    []assignmentScopedInlinePart `json:"parts"`
}

// One read may include several small first parts from the same selected source.
// Larger values keep their original continuation; no part is silently cut.
func (s Server) assignmentScopedInline(ctx context.Context, snapshot AssignmentContextSnapshot, resource AssignmentScopedResource, view string) (any, []AssignmentContextSource, bool, error) {
	out := assignmentScopedInlineView{Resource: resource, Parts: []assignmentScopedInlinePart{}}
	out.Resource.Parts = nil
	sources := []AssignmentContextSource{readScopeSource(resource, "title", "")}
	for _, manifest := range resource.Parts {
		if view == AssignmentContextViewComments && manifest.Part != AssignmentContentComments {
			continue
		}
		out.Parts = append(out.Parts, assignmentScopedInlinePart{Manifest: manifest})
	}
	for i := range out.Parts {
		manifest := out.Parts[i].Manifest
		if len(manifest.Cursors) == 0 {
			continue
		}
		value, pageSources, ok, err := s.assignmentContentView(ctx, snapshot, AgentAssignmentContextRequest{View: AssignmentContextViewContent, Target: resource.ID, Cursor: manifest.Cursors[0]})
		if err != nil || !ok {
			return nil, nil, false, err
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, nil, false, taskContextReadError{"invalid"}
		}
		var page AssignmentContentPage
		if json.Unmarshal(encoded, &page) != nil {
			return nil, nil, false, taskContextReadError{"invalid"}
		}
		out.Parts[i].Page = &page
		if len(manifest.Cursors) > 1 {
			out.Parts[i].NextCursor = manifest.Cursors[1]
		}
		trial, err := json.Marshal(out)
		if err != nil {
			return nil, nil, false, taskContextReadError{"invalid"}
		}
		if len(trial) > assignmentScopedInlineBytes {
			out.Parts[i].Page = nil
			out.Parts[i].NextCursor = ""
			continue
		}
		out.Parts[i].Manifest.NextAction = ""
		if out.Parts[i].NextCursor != "" {
			out.Parts[i].Manifest.NextAction = "riido context part " + resource.ID + " --cursor " + out.Parts[i].NextCursor
		}
		sources = append(sources, pageSources...)
	}
	return out, sources, true, nil
}
