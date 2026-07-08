package main

import (
	"strings"
	"testing"
)

func TestVerifyRejectsInvalidHeader(t *testing.T) {
	t.Parallel()
	m := testManifest()
	m.Title = ""
	err := verifyHeader(m)
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("verifyHeader error = %v, want required field", err)
	}
}

func TestVerifyDomainRejectsMissingSemanticAnchors(t *testing.T) {
	t.Parallel()
	cases := map[string]func(*manifest){
		"surface":        func(m *manifest) { m.Surfaces = m.Surfaces[1:] },
		"resource":       func(m *manifest) { m.Resources = m.Resources[1:] },
		"contract":       func(m *manifest) { m.ExternalContractVersions = m.ExternalContractVersions[1:] },
		"runtime config": func(m *manifest) { m.RuntimeConfigKeys = nil },
		"transport":      func(m *manifest) { m.TokenTransports = m.TokenTransports[1:] },
	}
	for name, mutate := range cases {
		m := testManifest()
		mutate(&m)
		if err := verifyDomain(m); err == nil || !strings.Contains(err.Error(), name) {
			t.Fatalf("%s verifyDomain error = %v", name, err)
		}
	}
}

func TestVerifyRuleGroupsRejectsMissingRule(t *testing.T) {
	t.Parallel()
	m := testManifest()
	m.RuleGroups[0].Rules = nil
	if err := verifyRuleGroups(m.RuleGroups); err == nil || !strings.Contains(err.Error(), "rule") {
		t.Fatalf("verifyRuleGroups error = %v, want missing rule", err)
	}
}

func TestVerifyLoopRejectsMissingStep(t *testing.T) {
	t.Parallel()
	m := testManifest()
	m.Loop.Execute = ""
	if err := verifyLoop(m.Loop); err == nil || !strings.Contains(err.Error(), "execute") {
		t.Fatalf("verifyLoop error = %v, want execute", err)
	}
}
