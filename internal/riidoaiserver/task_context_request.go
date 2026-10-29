package riidoaiserver

import "context"

type AIAgentTaskContextRequest struct {
	ComponentID       string
	WorkspaceID       string
	BearerToken       string
	DocumentSnapshot  *AIAgentTaskDocumentSnapshot
	Extended          bool // also read related work, meetings and comments (extra upstream reads)
	ReplyMetadata     bool // current-viewer audit only; never adds reply bodies to assignment context
	capturedReadScope bool // server-selected current-viewer read for an existing scoped snapshot
}

type AIAgentTaskContextRequestReader interface {
	GetAIAgentTaskContextForRequest(ctx context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error)
}
