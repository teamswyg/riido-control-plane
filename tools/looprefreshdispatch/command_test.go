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

func TestParseRefreshWorkflowDispatchAcceptsSafeInputs(t *testing.T) {
	root := repoRootForTest(t)
	workflow, verified, inputs, err := parseRefreshWorkflowDispatch(
		root,
		"gh workflow run ai-agent-client-testnet-load.yml -f scenario=public -f duration=120s -f concurrency=128",
	)
	if err != nil {
		t.Fatal(err)
	}
	if workflow != "ai-agent-client-testnet-load.yml" {
		t.Fatalf("workflow = %q", workflow)
	}
	want := "gh workflow run ai-agent-client-testnet-load.yml --ref main -f scenario=public -f duration=120s -f concurrency=128"
	if verified != want {
		t.Fatalf("verified command = %q, want %q", verified, want)
	}
	if len(inputs) != 3 || inputs[0].Name != "scenario" || inputs[0].Value != "public" {
		t.Fatalf("inputs = %+v", inputs)
	}
}

func TestParseRefreshWorkflowCommandRejectsUnsafeCommand(t *testing.T) {
	root := repoRootForTest(t)
	for _, command := range []string{
		"gh workflow run ../deploy.yml --ref main",
		"gh workflow run loop-registry.yml --ref main; echo bad",
		"bash -c 'gh workflow run loop-registry.yml --ref main'",
		"gh workflow run missing.yml --ref main",
		"gh workflow run loop-registry.yml --ref dev",
		"gh workflow run loop-registry.yml -f scenario=$(bad)",
	} {
		if _, err := parseRefreshWorkflowCommand(root, command); err == nil {
			t.Fatalf("expected unsafe command rejection for %q", command)
		}
	}
}
