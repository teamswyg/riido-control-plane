package riidoaiserver

import "strings"

// Native selection is explicit. Authorization or lookup errors never retry a
// legacy locator, which could refer to a different migrated entity.
func nativeReadScopeQuery(query string) string {
	query = strings.Replace(query, "($", "($workspaceId:String!,$locatorVersion:Int!,$", 1)
	query = strings.Replace(query, "$teamId:String!", "$teamId:String", 1)
	query = strings.Replace(query, "teamId:$teamId,", "workspaceId:$workspaceId,locatorVersion:$locatorVersion,teamId:$teamId,", 1)
	return strings.Replace(query, "schemaVersion rootId", "schemaVersion workspaceId rootId", 1)
}

func validReadScopeSchema(schema, workspace string) bool {
	return (schema == taskReadScopeSchema && workspace == "") || (schema == taskReadScopeNativeSchema && workspace != "")
}

func scopeMatchesLocator(scope taskReadScope, locator int, workspace string) bool {
	if locator == 2 {
		return scope.SchemaVersion == taskReadScopeNativeSchema && workspace != "" && scope.WorkspaceID == workspace
	}
	return scope.SchemaVersion == taskReadScopeSchema && scope.WorkspaceID == ""
}

func nativeReadScopeKind(resource AssignmentScopedResource) string {
	if resource.WorkspaceID != "" {
		return string(resource.Kind)
	}
	return ""
}

func (s Server) invocationReadScopeSchema() string {
	if reader, ok := s.taskContext.(*AIAgentPrivateTaskContextClient); ok && reader.locatorVersion == 2 {
		return taskReadScopeNativeSchema
	}
	if s.config.AssignmentReadScope {
		return taskReadScopeSchema
	}
	return ""
}
