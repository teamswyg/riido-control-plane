package riidoaiserver

import (
	"crypto/sha256"
	"fmt"
	"maps"
	"slices"
)

// CommentInvocation is created only from authenticated Monolith evidence. It is
// retained with the execution snapshot, never accepted from a client or daemon.
type CommentInvocation struct {
	EvidenceAgentID    string            `json:"evidence_agent_id,omitempty"`
	HumanFollowup      bool              `json:"human_followup,omitempty"`
	OriginAssignmentID string            `json:"origin_assignment_id,omitempty"`
	OriginThreadID     string            `json:"origin_thread_id,omitempty"`
	OriginMessageID    string            `json:"origin_message_id,omitempty"`
	EvidenceEventID    string            `json:"evidence_event_id"`
	RootCommentID      string            `json:"root_comment_id"`
	PrincipalID        string            `json:"principal_id"`
	WorkspaceID        string            `json:"workspace_id"`
	Ancestors          []string          `json:"ancestors"`
	Depth              int               `json:"depth"`
	Remaining          int               `json:"remaining"`
	Outcomes           map[string]string `json:"outcomes,omitempty"`
}

func newCommentInvocation(e verifiedCommentDelivery) *CommentInvocation {
	root := e.RootCommentID
	if root == "" {
		root = e.CommentID
	}
	return &CommentInvocation{EvidenceEventID: e.EventID, RootCommentID: root, PrincipalID: e.PrincipalID, WorkspaceID: e.WorkspaceID, Ancestors: []string{e.AgentID}, Remaining: 8}
}

func copyCommentInvocation(in *CommentInvocation) *CommentInvocation {
	if in == nil {
		return nil
	}
	out := *in
	out.Ancestors = slices.Clone(in.Ancestors)
	out.Outcomes = maps.Clone(in.Outcomes)
	return &out
}

func commentInvocationRequestID(assignmentID, target string) string {
	return fmt.Sprintf("comment-call:%x", sha256.Sum256([]byte(assignmentID+"\x00"+target)))
}
