package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (index *taskReadEvidenceIndex) read(ctx context.Context, root string, variables, out any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	v, ok := variables.(map[string]any)
	if !ok {
		return taskContextReadError{"invalid"}
	}
	stringValue := func(name string) string { value, _ := v[name].(string); return value }
	var key, kind, id string
	switch root {
	case "getComponentReadScope":
		if v["componentId"] != index.scope.RootID || v["teamId"] != index.scope.TeamID || v["editorSnapshot"] != nil {
			return taskContextReadError{"invalid"}
		}
		data, err := json.Marshal(index.scope)
		if err != nil {
			return taskContextReadError{"invalid"}
		}
		return json.Unmarshal(data, out)
	case "getReadScopeComments":
		kind, id = stringValue("resourceKind"), stringValue("resourceId")
		key = readEvidenceKey("comments", stringValue("resourceKind"), stringValue("resourceId"), stringValue("parentCommentId"), v["page"], stringValue("version"))
	case "getMeetingReadScope":
		kind, id = "meeting", stringValue("meetingId")
		key = readEvidenceKey("meeting", stringValue("meetingId"), stringValue("boardCursor"))
	default:
		switch root {
		case "getComponentReadScopeBody":
			kind, id = "component", stringValue("componentId")
		case "getBacklogReadScopeBody":
			kind, id = "backlog", stringValue("backlogId")
		case "getIntakeReadScopeBody":
			kind, id = "intake", stringValue("intakeId")
		case "getMeetingReadScopeBody":
			kind, id = "meeting", stringValue("meetingId")
		default:
			return taskContextReadError{"invalid"}
		}
		var selector *taskReadPartSelector
		if value := v["selector"]; value != nil {
			raw, err := json.Marshal(value)
			if err != nil || json.Unmarshal(raw, &selector) != nil {
				return taskContextReadError{"invalid"}
			}
		}
		var err error
		key, err = readEvidenceBodyKey(AssignmentContentKind(kind), id, stringValue("part"), stringValue("version"), v["offset"], selector)
		if err != nil {
			return err
		}
	}
	if team := index.selected[kind+"\x00"+id]; team == "" || team != stringValue("teamId") {
		return taskContextReadError{"invalid"}
	}
	data, found := index.values[key]
	if !found {
		if !index.complete {
			return errReadScopeCaptureIncomplete
		}
		return taskContextReadError{"invalid"}
	}
	if json.Unmarshal(data, out) != nil {
		return taskContextReadError{"invalid"}
	}
	return nil
}
