package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func nativeTestDigest(identity []string) string {
	raw, _ := json.Marshal(identity)
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func nativeTestBody(id, kind, team, content string, offset int, piece string, complete bool) map[string]any {
	page := readScopeCaptureBody(offset, piece, nil, complete)
	page["resourceId"], page["status"] = id, "available"
	page["version"] = nativeTestDigest([]string{"riido-read-body.v2", "workspace", kind, id, team, "document", "provider_json", "html", content})
	if !complete {
		page["nextOffset"] = offset + len(piece)
	}
	return page
}

func nativeTestWork(id, kind, team, reason string) map[string]any {
	work := readScopeCaptureInitial()["works"].([]any)[0].(map[string]any)
	work["id"], work["kind"], work["teamId"], work["workspaceId"] = id, kind, team, "workspace"
	work["reasons"] = []any{map[string]string{"kind": reason}}
	classification := "task"
	if reason == "project" {
		classification = "project"
		work["teamId"] = nil
	}
	work["summary"] = map[string]any{"id": id, "kind": kind, "title": "current " + id, "componentType": classification}
	detail := map[string]any{"id": id, "title": "current " + id, "componentType": classification, "status": map[string]string{"statusType": "todo"}}
	work["details"] = detail
	if kind == "backlog" {
		work["details"], work["backlogDetails"] = nil, detail
	}
	if kind == "intake" {
		work["details"], work["intakeDetails"] = nil, detail
	}
	work["body"] = nativeTestBody(id, kind, team, "helloworld", 0, "hello", false)
	comments := work["comments"].(map[string]any)
	comments["resourceId"], comments["resourceKind"] = id, kind
	return work
}

func nativeTestCapture(t *testing.T, kind string) *taskReadScopeEvidence {
	t.Helper()
	scope := readScopeCaptureInitial()
	scope["schemaVersion"], scope["workspaceId"] = taskReadScopeNativeSchema, "workspace"
	scope["works"] = []any{nativeTestWork("task-private", kind, "team", "root"), nativeTestWork("project", "component", "", "project")}
	pages := []any{
		map[string]any{"resourceKind": kind, "page": nativeTestBody("task-private", kind, "team", "helloworld", 5, "world", true)},
		map[string]any{"resourceKind": "component", "page": nativeTestBody("project", "component", "", "helloworld", 5, "world", true)},
	}
	raw, _ := json.Marshal(map[string]any{"schemaVersion": "riido-read-scope-capture.v1", "scope": scope, "bodyPages": pages, "commentPages": []any{}, "meetingPages": []any{}, "complete": true})
	var result taskReadScopeEvidence
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return &result
}
