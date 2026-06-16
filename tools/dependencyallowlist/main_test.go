package main

import (
	"os"
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

func TestLoadContractRejectsUnapprovedEntry(t *testing.T) {
	path := writeContract(t, `{
		"schema_version": "riido-go-dependency-allowlist.v2",
		"service": "riido-control-plane",
		"policy": "test",
		"allowed_direct_modules": [
			{
				"path": "go.opentelemetry.io/otel",
				"layer": "observability",
				"owner": "platform",
				"approved": false,
				"reason": "test"
			}
		]
	}`)
	_, err := loadContract(path)
	if err == nil || !strings.Contains(err.Error(), "approved must be true") {
		t.Fatalf("expected approved flag error, got %v", err)
	}
}

func TestLoadContractRejectsUnknownLayer(t *testing.T) {
	path := writeContract(t, `{
		"schema_version": "riido-go-dependency-allowlist.v2",
		"service": "riido-control-plane",
		"policy": "test",
		"allowed_direct_modules": [
			{
				"path": "github.com/example/config-framework",
				"layer": "framework",
				"owner": "platform",
				"approved": true,
				"reason": "test"
			}
		]
	}`)
	_, err := loadContract(path)
	if err == nil || !strings.Contains(err.Error(), "not in vocabulary") {
		t.Fatalf("expected layer vocabulary error, got %v", err)
	}
}

func testContract() contract {
	return contract{
		SchemaVersion: schemaVersion,
		Service:       "riido-control-plane",
		Policy:        "test",
		AllowedDirectModules: []allowedModule{
			{Path: "github.com/teamswyg/riido-contracts", Layer: "contract", Owner: "platform", Approved: true, Reason: "test"},
			{Path: "go.opentelemetry.io/otel", Layer: "observability", Owner: "platform", Approved: true, Reason: "test"},
			{Path: "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp", Layer: "observability", Owner: "platform", Approved: true, Reason: "test"},
			{Path: "go.opentelemetry.io/otel/sdk", Layer: "observability", Owner: "platform", Approved: true, Reason: "test"},
			{Path: "go.opentelemetry.io/otel/trace", Layer: "observability", Owner: "platform", Approved: true, Reason: "test"},
		},
	}
}

func writeContract(t *testing.T, data string) string {
	t.Helper()
	path := t.TempDir() + "/dependency_allowlist.riido.json"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	return path
}
