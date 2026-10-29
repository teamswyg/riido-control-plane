package riidoaiserver

import (
	"context"
	"errors"
)

func readScopeIncomplete(ctx context.Context, err error) bool {
	return ctx.Err() != nil || errors.Is(err, errReadScopeCaptureIncomplete)
}

func (b *taskReadScopeBuilder) missingPart(resource AssignmentScopedResource, part AssignmentContentPart, status string) AssignmentContentManifest {
	manifest := b.addPart(resource, part, "", nil, false)
	manifest.Status = status
	b.capture.Manifest[len(b.capture.Manifest)-1] = manifest
	return manifest
}

func (b *taskReadScopeBuilder) missingMeeting(ref taskReadScopeMeetingRef) AssignmentScopedResource {
	resource := AssignmentScopedResource{Kind: AssignmentContentMeeting, ID: ref.ID, TeamID: ref.TeamID, Reasons: ref.Reasons}
	for _, part := range []AssignmentContentPart{AssignmentContentMetadata, AssignmentContentEditor, AssignmentContentDescription, AssignmentContentBoard} {
		resource.Parts = append(resource.Parts, b.missingPart(resource, part, "unavailable"))
	}
	return resource
}
