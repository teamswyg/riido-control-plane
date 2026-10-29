package riidoaiserver

import "testing"

func TestCoverageRuntimeScopeIncludesEveryApprovedFallback(t *testing.T) {
	for _, tc := range []struct {
		provider, fallback, want string
	}{
		{"codex", "cursor", "local_invocations_quiescent"},
		{"claude", "codex", "local_invocations_quiescent"},
		{"cursor", "claude", "local_invocations_quiescent"},
		{"openclaw", "codex", "unobserved"},
		{"codex", "openclaw", "unobserved"},
		{"codex", "future-provider", "unobserved"},
		{"", "codex", "unobserved"},
	} {
		t.Run(tc.provider+"_"+tc.fallback, func(t *testing.T) {
			policy := recoveryTestPolicy()
			policy.Fallbacks[0].Provider = tc.fallback
			policy.Fallbacks[0].ModelID = "fixture-model"
			if got := coverageTerminationScope(Assignment{RuntimeProvider: tc.provider, RecoveryPolicy: policy}); got != tc.want {
				t.Fatalf("scope=%s want=%s", got, tc.want)
			}
		})
	}
}
