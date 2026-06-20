package main

import (
	"strings"
	"testing"
)

func TestVerifyModulesAllowsApprovedDirectDependencies(t *testing.T) {
	report, err := verifyModules(testContract(), testModules())
	if err != nil {
		t.Fatalf("verifyModules: %v", err)
	}
	if !strings.Contains(report, "verified 5 approved direct Go dependencies") {
		t.Fatalf("report = %q", report)
	}
}

func TestVerifyModulesRejectsUnapprovedDirectDependency(t *testing.T) {
	modules := append(testModules(), goModule{Path: "github.com/example/config-framework", Version: "v1.0.0"})
	_, err := verifyModules(testContract(), modules)
	if err == nil || !strings.Contains(err.Error(), "github.com/example/config-framework") {
		t.Fatalf("expected unapproved direct dependency error, got %v", err)
	}
}

func TestVerifyModulesRejectsUnusedAllowlistEntry(t *testing.T) {
	_, err := verifyModules(testContract(), []goModule{
		{Path: "github.com/teamswyg/riido-control-plane", Main: true},
		{Path: "github.com/teamswyg/riido-contracts", Version: "v0.3.6"},
	})
	if err == nil || !strings.Contains(err.Error(), "go.opentelemetry.io/otel") {
		t.Fatalf("expected unused allowlist entry error, got %v", err)
	}
}
