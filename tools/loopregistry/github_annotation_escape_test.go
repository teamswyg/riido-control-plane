package main

import "testing"

func TestGitHubAnnotationEscapesWorkflowSyntax(t *testing.T) {
	got := githubAnnotationProperty("title: one,two")
	if got != "title%3A one%2Ctwo" {
		t.Fatalf("property escape = %q", got)
	}
	if got := githubAnnotationMessage("a%b\nc"); got != "a%25b%0Ac" {
		t.Fatalf("message escape = %q", got)
	}
}
