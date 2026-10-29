package riidoaiserver

import "context"

type capturedScopeViewerReader struct{ requests []AIAgentTaskContextRequest }

func (r *capturedScopeViewerReader) GetAIAgentTaskContext(context.Context, string) (AIAgentTaskContext, error) {
	return AIAgentTaskContext{}, ErrAuthorizationForbidden
}

func (r *capturedScopeViewerReader) GetAIAgentTaskContextForRequest(_ context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	r.requests = append(r.requests, req)
	if req.capturedReadScope {
		return scopedViewerContext(), nil
	}
	return AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task"}}, nil
}

func installCapturedScopeViewer(s *Server) *capturedScopeViewerReader {
	reader := &capturedScopeViewerReader{}
	s.taskContext = reader
	store := s.config.AssignmentContext.(*MemoryAssignmentContextStore)
	store.mu.Lock()
	store.snapshots["assignment"] = AssignmentContextSnapshot{AssignmentID: "assignment", TaskID: "task", WorkspaceID: defaultAIAgentClientWorkspaceID, Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task"}, ReadScope: &AssignmentReadScope{SchemaVersion: taskReadScopeSchema, RootID: "task"}}}
	store.mu.Unlock()
	return reader
}
