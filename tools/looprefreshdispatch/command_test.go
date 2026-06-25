package main

import "testing"

func TestParseRefreshWorkflowCommandAcceptsExistingWorkflow(t *testing.T) {
	root := repoRootForTest(t)
	got, err := parseRefreshWorkflowCommand(
		root,
		"gh workflow run loop-registry.yml --ref main",
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != "loop-registry.yml" {
		t.Fatalf("workflow = %q", got)
	}
}

func TestParseRefreshWorkflowCommandRejectsUnsafeCommand(t *testing.T) {
	root := repoRootForTest(t)
	for _, command := range []string{
		"gh workflow run ../deploy.yml --ref main",
		"gh workflow run loop-registry.yml --ref main; echo bad",
		"bash -c 'gh workflow run loop-registry.yml --ref main'",
		"gh workflow run missing.yml --ref main",
	} {
		if _, err := parseRefreshWorkflowCommand(root, command); err == nil {
			t.Fatalf("expected unsafe command rejection for %q", command)
		}
	}
}
