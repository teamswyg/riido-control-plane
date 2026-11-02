package riidoaiserver

import "context"

type AssignmentContextStore interface {
	ConsumeAssignmentContextRead(ctx context.Context, assignmentID string) (bool, error)
	SaveAssignmentContextSnapshot(ctx context.Context, snapshot AssignmentContextSnapshot) error
	LoadAssignmentContextSnapshot(ctx context.Context, assignmentID string) (AssignmentContextSnapshot, bool, error)
	AppendAssignmentContextReads(ctx context.Context, reads []AssignmentContextRead) error
	ListAssignmentContextReads(ctx context.Context, assignmentID string) ([]AssignmentContextRead, error)
	SaveAssignmentContextComments(ctx context.Context, assignmentID string, pages [][]AIAgentTaskContextComment) error
	LoadAssignmentContextCommentPage(ctx context.Context, assignmentID string, page int) ([]AIAgentTaskContextComment, bool, error)
}
