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
	case "moved team":
		capture.Scope.TeamID = "moved-team"
		capture.Scope.Works[0].TeamID = "moved-team"
	case "downgrade":
		capture.Scope.SchemaVersion = taskReadScopeSchema
	}
}

func nativeTestLocation() privateComponentWorkspaceResponse {
	var out privateComponentWorkspaceResponse
	out.ID, out.ComponentType = "task-private", "task"
	out.Team.ID, out.Team.Workspace.ID = "team", "workspace"
	return out
}

func mutateNativeLocation(name string, location *privateComponentWorkspaceResponse) {
	switch name {
	case "location root":
		location.ID = "other"
	case "location workspace":
		location.Team.Workspace.ID = "other"
	case "location team":
		location.Team.ID = ""
	case "location classification":
		location.ComponentType = "project"
	}
}
