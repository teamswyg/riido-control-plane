package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestExplicitAssignmentNegotiatesConfiguredReadScope(t *testing.T) {
	for _, native := range []bool{false, true} {
		name, want := "legacy", taskReadScopeSchema
		if native {
			name, want = "native", taskReadScopeNativeSchema
		}
		t.Run(name, func(t *testing.T) {
			calls := 0
			f := explicitReadFixture(t, func(_ *assignmentContextFixture, v *explicitAssignmentResult, w http.ResponseWriter, r *http.Request) bool {
				calls++
				if r.Header.Get(commentReadScopeVersionHeader) != want {
					t.Errorf("read scope header = %q, want %q", r.Header.Get(commentReadScopeVersionHeader), want)
					w.WriteHeader(http.StatusBadRequest)
					return false
				}
				if native {
					v.SchemaVersion, v.Kind, v.TeamID = "riido-agent-explicit-read.v2", "project", nil
				}
				return true
			})
			if native {
				private, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{
					BaseURL: f.service.config.CommentDelivery.EvidenceBaseURL, LocatorVersion: 2,
				})
				if err != nil {
					t.Fatal(err)
				}
				f.service.taskContext = private
				f.server = f.service.Handler()
			}
			code, body := f.read("task", "target-one", f.leased.LeaseToken)
			if code != http.StatusOK || calls != 1 {
				t.Fatalf("status=%d calls=%d body=%s", code, calls, body)
			}
		})
	}
}
