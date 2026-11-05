package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCommentMentionPrecheckUsesCurrentTaskWorktreeOncePerDirectoryRequest(t *testing.T) {
	s, source, target, principal, _ := nativeResumeScopeFixture(t)
	client := s.aiAgent.(*DevelopmentAIAgentClientStore)
	current := Assignment{ID: "parent", TaskID: target.TaskID, AgentID: "agent-public-openclaw", WorkspaceID: principal.WorkspaceID, CreatedBy: principal.PrincipalID, Worktree: &AssignmentWorktree{RepositoryFullName: "old/repository", BranchName: "old-branch"}}
	evidence := verifiedCommentDelivery{EventID: "event", CommentID: "root", RootCommentID: "root", TaskID: current.TaskID, AgentID: current.AgentID, WorkspaceID: current.WorkspaceID, PrincipalID: current.CreatedBy, Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: current.TaskID, ComponentType: "task", Title: "Review current repository", BranchName: "current-branch"}, Repositories: []AIAgentTaskContextRepository{{FullName: "teamswyg/riido-daemon", RepositoryURL: "https://github.com/teamswyg/riido-daemon", Source: TaskContextRepositorySourceConnectedPullRequest}}}}
	source.projection.Assignment.Worktree = assignmentWorktreeFromTaskContext(evidence.Context, evidence.Context.Repositories[0])
	target.ConversationID = "comment:root"
	target.AssignmentState, target.WorkStatus = AgentAssignmentStateCompleted, AgentWorkStatusCompleted
	target.StartedAt = time.Now().Add(-time.Hour)
	client.taskThreads[current.TaskID] = []AIAgentTaskThreadRecord{{ThreadID: "parent-thread", AssignmentID: current.ID, TaskID: current.TaskID, AgentID: current.AgentID, ConversationID: "comment:root", CommentInvocation: &CommentInvocation{EvidenceEventID: "event", RootCommentID: "root", PrincipalID: current.CreatedBy, WorkspaceID: current.WorkspaceID, Ancestors: []string{current.AgentID}, Remaining: 8}}, target}
	var reads atomic.Int32
	key := strings.Repeat("k", 32)
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		if r.Header.Get(CommentDeliveryKeyHeader) != key {
			t.Error("missing evidence credential")
		}
		if err := json.NewEncoder(w).Encode(evidence); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(endpoint.Close)
	s.config.CommentDelivery = CommentDeliveryConfig{Key: key, EvidenceBaseURL: endpoint.URL}
	s.config.AgentMentionPrecheck = s.precheckCommentMention
	read := func() AgentDirectoryEntry {
		t.Helper()
		data, _, _, err := s.assignmentAgentDirectoryView(t.Context(), current, false, AgentAssignmentContextRequest{View: AssignmentContextViewAgents}, time.Now())
		page, _ := data.(AgentDirectoryPage)
		if err != nil || !page.PolicyChecked {
			t.Fatalf("directory: %+v %v", page, err)
		}
		for _, entry := range page.Agents {
			if entry.AgentID == target.AgentID {
				return entry
			}
		}
		t.Fatal("missing target")
		return AgentDirectoryEntry{}
	}
	if entry := read(); !entry.Callable {
		t.Fatalf("current compatible worktree rejected using parent's old worktree: %+v", entry)
	}
	if reads.Load() != 1 {
		t.Fatalf("one list fetched evidence %d times", reads.Load())
	}
	evidence.Context.Component.BranchName = "changed-again"
	if entry := read(); entry.Callable || !slices.Contains(entry.UnavailableReasons, "session_conflict") {
		t.Fatalf("next directory reused stale evidence: %+v", entry)
	}
	if reads.Load() != 2 {
		t.Fatalf("next request did not refetch evidence: %d", reads.Load())
	}
}

func TestCommentMentionPrecheckDoesNotClaimCheckedWhenEvidenceReadFails(t *testing.T) {
	f := newCommentCallFixture(t)
	projection, _, _ := f.assignments.LoadAssignmentProjection(t.Context(), f.source.AssignmentID)
	f.status = http.StatusServiceUnavailable
	_, _, _, err := f.s.assignmentAgentDirectoryView(t.Context(), projection.Assignment, false, AgentAssignmentContextRequest{View: AssignmentContextViewAgent, Target: "agent-owned-codex"}, time.Now())
	if err == nil {
		t.Fatal("unavailable evidence was presented as checked")
	}
	if len(f.threads(t)) != 1 {
		t.Fatal("failed lookup published work")
	}
}
