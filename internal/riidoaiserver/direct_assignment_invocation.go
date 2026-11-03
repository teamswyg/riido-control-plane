package riidoaiserver

import (
	"context"
	"net/url"
	"strings"
)

// A direct conversation is owned by the control plane. Its root is an execution
// thread, not an ID in the monolith's human comments table.
type directAssignmentRoot struct {
	ThreadID     string `json:"thread_id"`
	AssignmentID string `json:"assignment_id"`
	TaskID       string `json:"task_id"`
	TeamID       string `json:"team_id"`
}

func newDirectInvocation(action AIAgentTaskActionResponse, principal AuthorizationResult, team string) *CommentInvocation {
	return &CommentInvocation{DirectRoot: &directAssignmentRoot{ThreadID: action.ThreadID, AssignmentID: action.AssignmentID, TaskID: action.TaskID, TeamID: team}, PrincipalID: principal.PrincipalID, WorkspaceID: principal.WorkspaceID, EvidenceAgentID: action.AgentID, Ancestors: []string{action.AgentID}, Remaining: 8}
}

func (c *CommentInvocation) matchesConversation(t AIAgentTaskThreadRecord) bool {
	if root := c.DirectRoot; root != nil {
		return root.TeamID != "" && root.AssignmentID != "" && root.ThreadID != "" && root.TaskID == t.TaskID && root.ThreadID == taskThreadConversationID(t) && c.RootCommentID == "" && c.EvidenceEventID == ""
	}
	return c.RootCommentID != "" && t.ConversationID == "comment:"+c.RootCommentID
}

func (s Server) prepareDirectAssignment(ctx context.Context, p AuthorizationResult, task string, req *AssignAIAgentTaskRequest, composed *composedAssignRequest) error {
	if s.config.CommentDelivery.Key == "" || composed.TaskContext == nil || composed.Request.TeamID == "" || req.SourceCommentID != "" {
		return nil
	}
	if err := s.appendCommentInvocationPrompt(ctx, p, task, &composed.Request); err != nil {
		return err
	}
	req.directRootTeamID = composed.Request.TeamID
	return nil
}

func (s Server) readInvocationEvidence(ctx context.Context, task string, c *CommentInvocation) (verifiedCommentDelivery, error) {
	root := c.DirectRoot
	if root == nil {
		return s.readCommentDeliveryEvidence(ctx, c.EvidenceEventID)
	}
	if root.TaskID != task || root.TeamID == "" || root.ThreadID == "" || root.AssignmentID == "" || c.PrincipalID == "" || c.WorkspaceID == "" || c.EvidenceAgentID == "" {
		return verifiedCommentDelivery{}, ErrAuthorizationForbidden
	}
	path := strings.TrimRight(s.config.CommentDelivery.EvidenceBaseURL, "/") + "/internal/agent-assignment-context/" + url.PathEscape(c.WorkspaceID) + "/" + url.PathEscape(root.TeamID) + "/" + url.PathEscape(task) + "/" + url.PathEscape(c.PrincipalID)
	evidence, err := s.fetchInvocationEvidence(ctx, path)
	if err != nil {
		return verifiedCommentDelivery{}, err
	}
	if evidence.Context.TeamID != root.TeamID || evidence.TaskID != task || evidence.PrincipalID != c.PrincipalID || evidence.WorkspaceID != c.WorkspaceID {
		return verifiedCommentDelivery{}, ErrAuthorizationForbidden
	}
	evidence.directConversationID = root.ThreadID
	evidence.AgentID = c.EvidenceAgentID
	return evidence, nil
}
