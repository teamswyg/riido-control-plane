package riidoaiserver

import "testing"

func TestNativeFollowupUsesAgentOwnershipSeparatelyFromRuntimeOwnership(t *testing.T) {
	for _, owner := range []string{"user-2", ""} {
		t.Run("current-owner-"+owner, func(t *testing.T) {
			s, _, thread, principal, req := nativeResumeScopeFixture(t)
			thread.AgentSnapshot.OwnerPrincipalID = "user-2"
			principal.PrincipalID, req.CreatedBy = "user-2", "user-2"
			err := s.applyThreadNativeResume(t.Context(), principal, thread, &req, owner)
			if owner == "" {
				if err == nil || req.ResumeSessionID != "" {
					t.Fatal("missing Agent ownership was accepted")
				}
			} else if err != nil || req.ResumeSessionID != "selected-native" {
				t.Fatal("connected account Agent owner was confused with physical runtime owner")
			}
		})
	}
}
