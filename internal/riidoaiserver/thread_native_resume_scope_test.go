package riidoaiserver

import "testing"

func TestNativeFollowupSelectsOnlyAuthorizedPersistedSession(t *testing.T) {
	for _, name := range []string{"completed", "failed", "cancelled", "authorized-different-requester", "changed-model", "changed-provider", "changed-worktree", "foreign-task", "foreign-workspace", "foreign-agent", "foreign-assignment", "owner-changed", "running-no-resume", "legacy-no-pin", "missing-source", "unsupported-openclaw", "existing-resume-fallback"} {
		t.Run(name, func(t *testing.T) {
			s, source, thread, principal, req := nativeResumeScopeFixture(t)
			wantResume, wantErr := "selected-native", false
			switch name {
			case "failed":
				source.projection.Assignment.State = AssignmentFailed
			case "cancelled":
				source.projection.Assignment.State = AssignmentCancelled
			case "authorized-different-requester":
				principal.PrincipalID = "user-2"
				req.CreatedBy = "user-2"
			case "changed-model":
				req.ModelID = "different-model"
				wantErr = true
			case "changed-provider":
				source.projection.Assignment.RuntimeProvider = "claude"
				wantErr = true
			case "changed-worktree":
				source.projection.Assignment.Worktree = &AssignmentWorktree{RepositoryURL: "https://example.invalid/foreign"}
				wantErr = true
			case "foreign-task":
				source.projection.Assignment.TaskID = "foreign-task"
				wantErr = true
			case "foreign-workspace":
				source.projection.Assignment.WorkspaceID = "foreign-workspace"
				wantErr = true
			case "foreign-agent":
				source.projection.Assignment.AgentID = "foreign-agent"
				wantErr = true
			case "foreign-assignment":
				source.projection.Assignment.ID = "not-selected-assignment"
				wantErr = true
			case "owner-changed":
				thread.AgentSnapshot.OwnerPrincipalID = "other-owner"
				wantErr = true
			case "running-no-resume":
				source.projection.Assignment.State = AssignmentRunning
				wantResume = ""
			case "legacy-no-pin":
				source.projection.Assignment.ProviderSessionID = ""
				wantResume = ""
			case "missing-source":
				source.found = false
				wantResume = ""
			case "unsupported-openclaw":
				req.RuntimeProvider = "openclaw"
				wantResume = ""
			case "existing-resume-fallback":
				source.projection.Assignment.ProviderSessionID = ""
				source.projection.Assignment.ResumeSessionID = "selected-native"
			}
			err := s.applyThreadNativeResume(t.Context(), principal, thread, &req)
			if wantErr {
				if err == nil || req.ResumeSessionID != "" {
					t.Fatal("invalid native scope was accepted")
				}
				return
			}
			if err != nil || req.ResumeSessionID != wantResume {
				t.Fatal("native session selection violated scope or legacy behavior")
			}
		})
	}
}
