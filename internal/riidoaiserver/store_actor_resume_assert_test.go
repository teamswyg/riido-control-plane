package riidoaiserver

import "testing"

func assertResumeAssignment(t *testing.T, assignment Assignment) {
	t.Helper()
	if assignment.ResumeSessionID != "sess-prev" {
		t.Fatalf("assignment resume_session_id = %q", assignment.ResumeSessionID)
	}
	if assignment.Worktree == nil ||
		assignment.Worktree.RepositoryFullName != "teamswyg/riido-daemon" ||
		assignment.Worktree.RepositoryURL != "https://github.com/teamswyg/riido-daemon" ||
		assignment.Worktree.BranchName != "RIID-4964-agent-profile-upload" ||
		assignment.Worktree.Source != "connected_pull_request" {
		t.Fatalf("assignment worktree was not normalized: %+v", assignment.Worktree)
	}
}
