package riidoaiserver

import (
	"encoding/json"
	"net/http"
)

type explicitAssignmentResult struct {
	SchemaVersion string          `json:"schema_version"`
	WorkspaceID   string          `json:"workspace_id"`
	RootID        string          `json:"root_id"`
	PrincipalID   string          `json:"principal_id"`
	ID            string          `json:"id"`
	TeamID        string          `json:"team_id"`
	Title         string          `json:"title"`
	Kind          string          `json:"kind"`
	Admin         bool            `json:"admin"`
	Truncated     bool            `json:"truncated"`
	Data          json.RawMessage `json:"data"`
}

func explicitAssignmentView(view string) bool {
	return view == "activity" || view == "threads" || view == "thread" || view == "body"
}

func validExplicitOptions(req AgentAssignmentContextRequest) bool {
	if req.Cursor != "" || len(req.Options) > 8 || len(req.Target) > 128 || len(req.ThreadID) > 512 {
		return false
	}
	if req.View != "thread" && req.ThreadID != "" {
		return false
	}
	for k, v := range req.Options {
		if len(v) > 256 {
			return false
		}
		allowed := false
		switch k {
		case "page", "limit":
			allowed = req.View == "activity" || req.View == "comments" || req.View == "body"
		case "version", "comment-id":
			allowed = req.View == "comments" || req.View == "body"
		case "part", "offset", "comments-limit":
			allowed = req.View == "body"
		}
		if !allowed {
			return false
		}
	}
	return req.View == "task" || req.View == "comments" || explicitAssignmentView(req.View)
}

func explicitReadHTTPStatus(code int) int {
	status := 503
	switch code {
	case http.StatusBadRequest:
		status = 400
	case http.StatusForbidden, http.StatusNotFound, http.StatusGone:
		status = 403
	case http.StatusConflict:
		status = 412
	case http.StatusUnprocessableEntity:
		status = 422
	}
	return status
}
