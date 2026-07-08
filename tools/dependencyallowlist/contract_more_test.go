package main

import (
	"strings"
	"testing"
)

func TestVerifyContractRejectsRequiredFields(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		edit func(*contract)
		want string
	}{
		{"schema", func(c *contract) { c.SchemaVersion = "v0" }, "schema_version"},
		{"service", func(c *contract) { c.Service = " " }, "service is required"},
		{"policy", func(c *contract) { c.Policy = "" }, "policy is required"},
		{"id", func(c *contract) { c.ID = "" }, "id is required"},
		{"assertion", func(c *contract) { c.Assertions = []string{" "} }, "assertions[0]"},
		{"loop", func(c *contract) { c.Loop.Execute = "" }, "loop.execute"},
	}
	for _, tc := range cases {
		c := testContract()
		tc.edit(&c)
		if err := verifyContract(c); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s error = %v, want %q", tc.name, err, tc.want)
		}
	}
}

func TestVerifyAllowedModuleRejectsShapeErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		module allowedModule
		want   string
	}{
		{allowedModule{}, "path is required"},
		{allowedModule{Path: "x", Layer: "contract", Owner: "", Reason: "r", Approved: true}, "must include"},
		{allowedModule{Path: "x", Layer: "unknown", Owner: "o", Reason: "r", Approved: true}, "not in vocabulary"},
	}
	for _, tc := range cases {
		if err := verifyAllowedModule(0, tc.module, map[string]struct{}{}); err == nil ||
			!strings.Contains(err.Error(), tc.want) {
			t.Fatalf("error = %v, want %q", err, tc.want)
		}
	}
}

func TestVerifyAllowedModuleRejectsDuplicate(t *testing.T) {
	t.Parallel()
	seen := map[string]struct{}{"github.com/example/mod": {}}
	module := allowedModule{Path: "github.com/example/mod", Layer: "contract", Owner: "o", Reason: "r", Approved: true}
	if err := verifyAllowedModule(1, module, seen); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}
