package riidoaiserver

import "unicode/utf8"

func readScopeEditorContinuation(resource AssignmentScopedResource, first taskReadBodyPage, editor *AIAgentTaskDocumentSnapshot, offset int) (*taskReadBodyPage, error) {
	if err := editor.validate(); err != nil {
		return nil, err
	}
	if resource.Kind != AssignmentContentComponent || first.Part != "document" || first.Format != editor.ContentFormat || first.Version != readScopeBodyVersion(resource.Kind, resource.ID, first.Part, "client_snapshot", editor.ContentFormat, editor.Content) || offset < 0 || offset >= len(editor.Content) || !utf8.RuneStart(editor.Content[offset]) {
		return nil, taskContextReadError{"invalid"}
	}
	end := min(offset+8192, len(editor.Content))
	for end < len(editor.Content) && !utf8.RuneStart(editor.Content[end]) {
		end--
	}
	page := first
	page.Offset = &offset
	page.Content = editor.Content[offset:end]
	complete := end == len(editor.Content)
	page.Complete = &complete
	page.NextOffset = nil
	page.Status = "available"
	if !complete {
		page.NextOffset = &end
		page.Status = "partial"
	}
	return &page, nil
}
