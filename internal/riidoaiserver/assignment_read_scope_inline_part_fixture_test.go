package riidoaiserver

import (
	"encoding/json"
)

func inlineBoundaryPart(kind AssignmentContentKind, id string, part AssignmentContentPart, cursor, auditPart, data string) (AssignmentScopedResource, AssignmentContentPage) {
	resource := AssignmentScopedResource{Kind: kind, ID: id, TeamID: "team", Title: id, Reasons: []taskReadScopeReason{{Kind: "related"}}, Parts: []AssignmentContentManifest{{ResourceKind: kind, SourceID: id, Part: part, Version: "v1", Status: "available", Complete: true, Cursors: []string{cursor}, NextAction: "riido context part " + id + " --cursor " + cursor}}}
	if id == "task-private" && kind == AssignmentContentComponent {
		resource.Reasons = []taskReadScopeReason{{Kind: "root"}}
	}
	page := AssignmentContentPage{ResourceKind: kind, Cursor: cursor, SourceID: id, Part: part, Version: "v1", Status: "available", Complete: true, Data: json.RawMessage(data), Sources: []AssignmentContextSource{readScopeSource(resource, auditPart, "v1")}}
	return resource, page
}
