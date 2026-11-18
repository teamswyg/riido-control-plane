package riidoaiserver

import "encoding/json"

func validExplicitAssignmentLocator(raw []byte, v explicitAssignmentResult) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || len(fields["team_id"]) == 0 {
		return false
	}
	switch v.SchemaVersion {
	case "riido-agent-explicit-read.v1":
		return v.TeamID != nil && *v.TeamID != ""
	case "riido-agent-explicit-read.v2":
		return (v.Kind == "project" || v.Kind == "milestone") && v.TeamID == nil
	default:
		return false
	}
}
