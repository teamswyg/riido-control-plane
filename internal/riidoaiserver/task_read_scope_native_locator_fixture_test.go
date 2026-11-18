package riidoaiserver

func mutateNativeLocatorCase(name string, capture *taskReadScopeEvidence, req *AIAgentTaskContextRequest) {
	switch name {
	case "missing workspace":
		req.WorkspaceID = ""
	case "missing bearer":
		req.BearerToken = ""
	case "missing task":
		req.ComponentID = ""
	case "scope workspace":
		capture.Scope.WorkspaceID = "other"
	case "scope root":
		capture.Scope.RootID = "other"
	case "missing team":
		capture.Scope.TeamID = ""
		capture.Scope.Works[0].TeamID = ""
	case "missing root team":
		capture.Scope.Works[0].TeamID = ""
	case "old root team":
		capture.Scope.Works[0].TeamID = "old-team"
	case "downgrade":
		capture.Scope.SchemaVersion = taskReadScopeSchema
	}
}
