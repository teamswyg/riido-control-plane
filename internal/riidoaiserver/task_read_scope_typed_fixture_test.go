package riidoaiserver

import (
	"encoding/json"
	"strings"
)

func readScopeTypedCaptureFixture() (map[string]any, map[string]any, map[string]*taskReadBodyPage, map[string]string, *taskReadScopeEvidence) {
	scope, meetings := readScopeWholeFixture()
	scope["works"] = scope["works"].([]any)[:1]
	scope["meetings"] = scope["meetings"].([]any)[:1]
	work := scope["works"].([]any)[0].(map[string]any)
	meeting := meetings["meeting-0"].(map[string]any)
	big := strings.Repeat("original-large-한글", 5000)
	metadata := map[string]any{}
	for _, key := range []string{"details", "backlogDetails", "intakeDetails", "links", "github", "repositories", "figma", "priorities", "checklists", "attachments", "sprints", "origin"} {
		metadata[key] = work[key]
	}
	metadata["github"] = []any{map[string]string{"id": "saved-pr", "body": big}}
	comments := map[string]any{"resourceId": "task-private", "resourceKind": "component", "version": "logical-comments", "status": "available", "page": 1, "limit": 20, "complete": true, "nextPage": nil, "parentCommentId": nil, "items": []any{map[string]any{"id": "big-comment", "body": big, "replyCount": 0}}}
	board := map[string]any{"version": "logical-board", "offset": 0, "nextCursor": nil, "complete": true, "groups": []any{map[string]string{"id": "empty-group"}}, "cards": []any{map[string]string{"id": "big-card", "body": big}}}
	payloads := map[string]string{}
	prefixes := map[string]*taskReadBodyPage{}
	for _, tc := range []struct {
		id, kind, part string
		value          any
		selector       *taskReadPartSelector
	}{{"task-private", "component", "metadata", metadata, nil}, {"task-private", "component", "comments_page", comments, nil}, {"meeting-0", "meeting", "board_page", board, nil}} {
		raw, _ := json.Marshal(tc.value)
		selector, _ := normalizeReadScopePartSelector(AssignmentContentKind(tc.kind), tc.part, tc.selector)
		version := readScopePartVersion(AssignmentContentKind(tc.kind), tc.id, "team", tc.part, "riido_"+tc.part, string(raw), selector)
		prefixes[tc.part] = typedJSONFixture(tc.id, tc.part, version, string(raw), 0, &selector)
		payloads[tc.part] = string(raw)
	}
	work["summary"] = map[string]any{"id": "task-private", "kind": "component", "title": "root", "updatedAt": "now"}
	work["contentPage"] = prefixes["metadata"]
	work["details"] = nil
	work["comments"] = map[string]any{"resourceId": "task-private", "resourceKind": "component", "version": "logical-comments", "status": "deferred", "page": 1, "limit": 20, "complete": false, "nextPage": nil, "items": []any{}, "contentPage": prefixes["comments_page"]}
	meeting["board"] = map[string]any{"version": "logical-board", "offset": 0, "nextCursor": nil, "complete": false, "groups": []any{}, "cards": []any{}, "contentPage": prefixes["board_page"]}
	complete := true
	rawScope, _ := json.Marshal(scope)
	var typedScope taskReadScope
	_ = json.Unmarshal(rawScope, &typedScope)
	rawMeeting, _ := json.Marshal(meeting)
	var typedMeeting taskReadMeeting
	_ = json.Unmarshal(rawMeeting, &typedMeeting)
	evidence := &taskReadScopeEvidence{SchemaVersion: "riido-read-scope-capture.v1", Scope: &typedScope, Complete: &complete, BodyPages: []taskReadCapturedBody{}, CommentPages: []*taskReadCommentPage{}, MeetingPages: []taskReadCapturedMeeting{{Snapshot: &typedMeeting}}}
	return scope, meeting, prefixes, payloads, evidence
}
