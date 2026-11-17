package riidoaiserver

// The CLI manifest is not the producer read-scope wire. Keep its v1 shape
// (resources/parts) for existing daemons and record native producer identity
// separately from that consumer version.
func nativeScopeSourceSchema(scope taskReadScope) string {
	if scope.SchemaVersion == taskReadScopeNativeSchema {
		return scope.SchemaVersion
	}
	return ""
}

func validAssignmentScopeSchema(scope *AssignmentReadScope, workspace string) bool {
	if scope.SchemaVersion != taskReadScopeSchema {
		return false
	}
	if scope.SourceSchemaVersion == "" {
		return scope.WorkspaceID == ""
	}
	return scope.SourceSchemaVersion == taskReadScopeNativeSchema && workspace != "" && scope.WorkspaceID == workspace
}

func invalidAssignmentResourceWorkspace(scope *AssignmentReadScope, resource AssignmentScopedResource) bool {
	if scope.SourceSchemaVersion == taskReadScopeNativeSchema && resource.Kind != AssignmentContentMeeting {
		return resource.WorkspaceID != scope.WorkspaceID
	}
	return resource.WorkspaceID != "" && resource.WorkspaceID != scope.WorkspaceID
}
