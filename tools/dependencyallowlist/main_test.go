package main

import (
	"strings"
	"testing"
)

func TestVerifyModulesAllowsApprovedDirectDependencies(t *testing.T) {
	report, err := verifyModules(testContract(), []goModule{
		{Path: "github.com/teamswyg/riido-control-plane", Main: true},
		{Path: "github.com/teamswyg/riido-contracts", Version: "v0.3.6"},
		{Path: "go.opentelemetry.io/otel", Version: "v1.44.0"},
		{Path: "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp", Version: "v1.44.0"},
		{Path: "go.opentelemetry.io/otel/sdk", Version: "v1.44.0"},
		{Path: "go.opentelemetry.io/otel/trace", Version: "v1.44.0"},
		{Path: "google.golang.org/grpc", Version: "v1.81.1", Indirect: true},
	})
	if err != nil {
		t.Fatalf("verifyModules: %v", err)
	}
	if !strings.Contains(report, "verified 5 approved direct Go dependencies") {
		t.Fatalf("report = %q", report)
	}
}

func TestVerifyModulesRejectsUnapprovedDirectDependency(t *testing.T) {
	_, err := verifyModules(testContract(), []goModule{
		{Path: "github.com/teamswyg/riido-control-plane", Main: true},
		{Path: "github.com/teamswyg/riido-contracts", Version: "v0.3.6"},
		{Path: "go.opentelemetry.io/otel", Version: "v1.44.0"},
		{Path: "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp", Version: "v1.44.0"},
		{Path: "go.opentelemetry.io/otel/sdk", Version: "v1.44.0"},
		{Path: "go.opentelemetry.io/otel/trace", Version: "v1.44.0"},
		{Path: "github.com/example/config-framework", Version: "v1.0.0"},
	})
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

func testContract() contract {
	return contract{
		SchemaVersion: schemaVersion,
		Service:       "riido-control-plane",
		Policy:        "test",
		AllowedDirectModules: []allowedModule{
			{Path: "github.com/teamswyg/riido-contracts", Category: "riido-contract", Owner: "platform", Reason: "test"},
			{Path: "go.opentelemetry.io/otel", Category: "observability", Owner: "platform", Reason: "test"},
			{Path: "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp", Category: "observability", Owner: "platform", Reason: "test"},
			{Path: "go.opentelemetry.io/otel/sdk", Category: "observability", Owner: "platform", Reason: "test"},
			{Path: "go.opentelemetry.io/otel/trace", Category: "observability", Owner: "platform", Reason: "test"},
		},
	}
}
