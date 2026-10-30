package riidoaiserver

import "encoding/json"

func projectReadScopeWork(out *AIAgentTaskContext, resource AssignmentScopedResource, work *taskReadScopeWork) {
	if resource.Kind != AssignmentContentComponent {
		return
	}
	if resource.ID == out.ReadScope.RootID {
		var detail struct {
			ComponentType string `json:"componentType"`
			BranchName    string `json:"branchName"`
		}
		_ = json.Unmarshal(work.Details, &detail)
		if work.ContentPage != nil && work.Summary != nil {
			if work.Summary.ComponentType != nil {
				detail.ComponentType = *work.Summary.ComponentType
			}
			if work.Summary.BranchName != nil {
				detail.BranchName = *work.Summary.BranchName
			}
		}
		out.Component = AIAgentTaskContextComponent{ID: resource.ID, Title: resource.Title, KeyNumber: resource.KeyNumber, UpdatedAt: resource.UpdatedAt, ComponentType: detail.ComponentType, BranchName: detail.BranchName}
		out.Document = AIAgentTaskContextDocument{Content: work.Body.Content, ContentFormat: work.Body.Format, Source: work.Body.Source, Status: work.Body.Status}
		out.Repositories = readScopeLegacyRepositories(work.Repositories)
	}
	for _, reason := range resource.Reasons {
		ref := AIAgentTaskContextReference{ID: resource.ID, Title: resource.Title, KeyNumber: resource.KeyNumber, UpdatedAt: resource.UpdatedAt}
		switch reason.Kind {
		case "parentTask":
			out.Hierarchy.ParentTask = ref
		case "milestone":
			out.Hierarchy.Milestone = ref
		case "project":
			out.Hierarchy.Project = ref
		}
	}
}

func validateReadScopeEditor(resource AssignmentScopedResource, body *taskReadBodyPage, editor *AIAgentTaskDocumentSnapshot) error {
	// The producer uses the click snapshot to select references, but may return
	// the persisted body. Its identity, version and pages use normal scope checks.
	if body.Source != "client_snapshot" {
		return nil
	}
	if editor == nil {
		return taskContextReadError{"invalid"}
	}
	if body.Format != editor.ContentFormat || body.Part != "document" || body.Version != readScopeBodyVersion(resource.Kind, resource.ID, "document", "client_snapshot", editor.ContentFormat, editor.Content) || len(body.Content) > len(editor.Content) || editor.Content[:len(body.Content)] != body.Content {
		return taskContextReadError{"invalid"}
	}
	return nil
}
