package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (c *AIAgentPrivateTaskContextClient) captureReadScopeMeeting(ctx context.Context, bearer string, b *taskReadScopeBuilder, ref taskReadScopeMeetingRef) (AssignmentScopedResource, error) {
	first, err := c.fetchReadScopeMeeting(ctx, bearer, ref, "", nil)
	if err != nil {
		if readScopeIncomplete(ctx, err) && first != nil {
			return b.captureIncompleteTypedMeeting(ref, first), nil
		}
		return AssignmentScopedResource{}, err
	}
	var detail struct {
		ID        string
		Title     string
		UpdatedAt string
	}
	if json.Unmarshal(first.Details, &detail) != nil || detail.ID != ref.ID {
		return AssignmentScopedResource{}, taskContextReadError{"invalid"}
	}
	resource := AssignmentScopedResource{Kind: AssignmentContentMeeting, ID: ref.ID, TeamID: ref.TeamID, Title: detail.Title, UpdatedAt: detail.UpdatedAt, Reasons: ref.Reasons}
	metadata, err := readScopeMeetingMetadata(*first)
	if err != nil {
		return AssignmentScopedResource{}, taskContextReadError{"invalid"}
	}
	version := assignmentContextVersion(string(metadata))
	resource.Parts = append(resource.Parts, b.addPart(resource, AssignmentContentMetadata, version, readScopeJSONPages(metadata, 0, 0, []AssignmentContextSource{readScopeSource(resource, "metadata", version)}), readScopeMetadataComplete(first.PartsStatus)))
	for _, body := range []*taskReadBodyPage{first.Body, first.DescriptionBody, first.AlternateBody} {
		if body == nil {
			continue
		}
		part, err := c.captureReadScopeBody(ctx, bearer, b, resource, body, nil)
		if err != nil {
			return AssignmentScopedResource{}, err
		}
		resource.Parts = append(resource.Parts, part)
	}
	part, err := c.captureReadScopeBoard(ctx, bearer, b, resource, ref, first)
	if err != nil {
		return AssignmentScopedResource{}, err
	}
	resource.Parts = append(resource.Parts, part)
	return resource, nil
}

func readScopeMeetingMetadata(meeting taskReadMeeting) (json.RawMessage, error) {
	meeting.Board, meeting.Body, meeting.DescriptionBody, meeting.AlternateBody = nil, nil, nil, nil
	return json.Marshal(meeting)
}
