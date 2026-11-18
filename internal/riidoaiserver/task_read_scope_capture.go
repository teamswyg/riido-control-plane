package riidoaiserver

import "context"

func (c *AIAgentPrivateTaskContextClient) getTaskReadScope(ctx context.Context, req AIAgentTaskContextRequest, team string) (AIAgentTaskContext, error) {
	scope, err := c.loadTaskReadScope(ctx, req, team)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	b, err := newTaskReadScopeBuilder()
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	out := AIAgentTaskContext{TeamID: team, ReadScope: &AssignmentReadScope{SchemaVersion: taskReadScopeSchema, SourceSchemaVersion: nativeScopeSourceSchema(scope), WorkspaceID: scope.WorkspaceID, RootID: scope.RootID, TeamID: team, Status: "available", Resources: []AssignmentScopedResource{}}, Repositories: []AIAgentTaskContextRepository{}}
	for _, work := range scope.Works {
		resource, _ := readScopeWorkSummary(work)
		resource.WorkspaceID = scope.WorkspaceID
		var rootEditor *AIAgentTaskDocumentSnapshot
		if resource.ID == scope.RootID {
			rootEditor = req.DocumentSnapshot
		}
		if err := validateReadScopeEditor(resource, work.Body, rootEditor); err != nil {
			return AIAgentTaskContext{}, err
		}
		metadata, metadataErr := c.captureReadScopeWorkMetadata(ctx, req.BearerToken, b, resource, work)
		if metadataErr != nil {
			return AIAgentTaskContext{}, metadataErr
		}
		resource.Parts = append(resource.Parts, metadata)
		for _, body := range []*taskReadBodyPage{work.Body, work.DescriptionBody, work.AlternateBody} {
			if body == nil {
				continue
			}
			var editor *AIAgentTaskDocumentSnapshot
			if resource.ID == req.ComponentID {
				editor = req.DocumentSnapshot
			}
			part, err := c.captureReadScopeBody(ctx, req.BearerToken, b, resource, body, editor)
			if err != nil {
				return AIAgentTaskContext{}, err
			}
			resource.Parts = append(resource.Parts, part)
		}
		comments, err := c.captureReadScopeComments(ctx, req.BearerToken, b, resource, work.Comments)
		if err != nil {
			return AIAgentTaskContext{}, err
		}
		resource.Parts = append(resource.Parts, comments)
		projectReadScopeWork(&out, resource, work)
		out.ReadScope.Resources = append(out.ReadScope.Resources, resource)
	}
	for _, meeting := range scope.Meetings {
		resource, err := c.captureReadScopeMeeting(ctx, req.BearerToken, b, *meeting)
		if err != nil {
			if !readScopeIncomplete(ctx, err) {
				return AIAgentTaskContext{}, err
			}
			resource = b.missingMeeting(*meeting)
		}
		out.ReadScope.Resources = append(out.ReadScope.Resources, resource)
	}
	for _, part := range b.capture.Manifest {
		if !part.Complete {
			out.ReadScope.Status = "partial"
		}
	}
	out.readScopeCapture = &b.capture
	return out, nil
}
