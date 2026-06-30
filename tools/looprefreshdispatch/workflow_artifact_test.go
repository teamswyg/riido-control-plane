package main

import (
	"strings"
	"testing"
)

func TestLoopRefreshDispatchWorkflowSeparatesSampleAndLiveArtifacts(t *testing.T) {
	text := loopRefreshDispatchWorkflowText(t)
	for _, required := range []string{
		"-evidence-out out/sample/loop-refresh-dispatch-plan.json",
		"-candidate-out out/sample/loop-refresh-dispatch-closed-loop-candidates.json",
		"if: github.event_name == 'pull_request'",
		"cp out/sample/loop-refresh-dispatch-plan.json out/loop-refresh-dispatch-plan.json",
		"rm -f out/loop-refresh-dispatch-plan.json",
		"rm -f out/loop-refresh-dispatch-closed-loop-candidates.json",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("workflow must keep sample evidence away from live artifacts: missing %q", required)
		}
	}
}
