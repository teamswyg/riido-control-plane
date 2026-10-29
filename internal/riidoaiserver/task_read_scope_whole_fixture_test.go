package riidoaiserver

import "fmt"

func readScopeWholeFixture() (map[string]any, map[string]any) {
	scope := readScopeCaptureInitial()
	works := []any{}
	for i := 0; i < 25; i++ {
		id, reason := fmt.Sprintf("work-%d", i), "mention"
		switch {
		case i == 0:
			id, reason = "task-private", "root"
		case i <= 5:
			reason = "related"
		case i <= 15:
			reason = "child"
		}
		r := map[string]any{"kind": reason}
		if reason == "related" {
			r["relation"] = "relatedTo"
		}
		work := map[string]any{"id": id, "kind": "component", "teamId": "team", "reasons": []any{r}, "details": map[string]any{"id": id, "title": id, "keyNumber": fmt.Sprint(i), "componentType": "task", "estimate": 0, "createdBy": nil}, "body": readScopeWholeBody(id, AssignmentContentComponent, "document", "<p>Body</p>"), "descriptionBody": readScopeWholeBody(id, AssignmentContentComponent, "description", "<p>Description</p>"), "comments": map[string]any{"resourceId": id, "resourceKind": "component", "version": "comments-" + id, "status": "empty", "page": 1, "limit": 20, "nextPage": nil, "complete": true, "items": []any{}}, "links": []any{}, "github": []any{}, "repositories": []any{}, "figma": nil, "priorities": []any{}, "checklists": []any{}, "attachments": []any{}, "sprints": []any{}, "origin": nil, "partsStatus": []any{}}
		works = append(works, work)
	}
	refs := []any{}
	meetings := map[string]any{}
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("meeting-%d", i)
		refs = append(refs, map[string]any{"id": id, "teamId": "team", "reasons": []any{map[string]string{"kind": "meeting"}}})
		meetings[id] = map[string]any{"id": id, "teamId": "team", "status": "available", "details": map[string]any{"id": id, "title": id, "meetingParticipants": []any{map[string]string{"id": "empty-user"}}}, "labels": []any{}, "board": map[string]any{"version": "board-" + id, "offset": 0, "nextCursor": nil, "complete": true, "groups": []any{map[string]string{"id": "empty-group"}}, "cards": []any{map[string]string{"id": "card-" + id, "body": "<p>Card</p>"}}}, "body": readScopeWholeBody(id, AssignmentContentMeeting, "document", "<p>Meeting</p>"), "descriptionBody": readScopeWholeBody(id, AssignmentContentMeeting, "description", "<p>Details</p>"), "partsStatus": []any{}}
	}
	scope["works"], scope["meetings"] = works, refs
	return scope, meetings
}

func readScopeWholeBody(id string, kind AssignmentContentKind, part, content string) map[string]any {
	return map[string]any{"resourceId": id, "part": part, "status": "available", "source": "snapshot_html", "format": "html", "version": readScopeBodyVersion(kind, id, part, "snapshot_html", "html", content), "offset": 0, "content": content, "nextOffset": nil, "complete": true}
}
