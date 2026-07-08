package main

import "testing"

func TestVerifyRejectsShapeAndLoopDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"schema":  func(m *manifest) { m.SchemaVersion = "other" },
		"binding": func(m *manifest) { m.RiskEvidence = "" },
		"sources": func(m *manifest) { m.Sources = nil },
		"owners":  func(m *manifest) { m.Owners = nil },
		"figma":   func(m *manifest) { m.Figma = nil },
		"checks":  func(m *manifest) { m.Checks = nil },
		"loop":    func(m *manifest) { m.Loop.Execute = "" },
	} {
		t.Run(name, func(t *testing.T) {
			m := apiClientDeliveryFixture()
			mutate(&m)
			if _, err := verifyAll(t.TempDir(), m); err == nil {
				t.Fatalf("expected shape error")
			}
		})
	}
}

func TestVerifyRejectsSourceAndRiskDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"source-shape": func(m *manifest) { m.Sources[0].Path = "" },
		"source-file":  func(m *manifest) { m.Sources[0].Path = "missing.json" },
		"check-file":   func(m *manifest) { m.Checks[0].File = "missing.go" },
		"check-token":  func(m *manifest) { m.Checks[0].Contains = []string{"missing"} },
		"risk-file":    func(m *manifest) { m.RiskEvidence = "missing-risk.json" },
	} {
		t.Run(name, func(t *testing.T) {
			m := apiClientDeliveryFixture()
			repo := writeAPIClientDeliveryRepo(t, m)
			mutate(&m)
			if _, err := verifyAll(repo, m); err == nil {
				t.Fatalf("expected source or risk error")
			}
		})
	}
}

func TestVerifyRejectsRenderedPhraseDrift(t *testing.T) {
	m := apiClientDeliveryFixture()
	repo := writeAPIClientDeliveryRepo(t, m)
	m.Required = []string{"missing phrase"}
	if _, err := verifyAll(repo, m); err == nil {
		t.Fatalf("expected required phrase error")
	}
	m = apiClientDeliveryFixture()
	repo = writeAPIClientDeliveryRepo(t, m)
	m.Forbidden = []string{"API Client Delivery"}
	if _, err := verifyAll(repo, m); err == nil {
		t.Fatalf("expected forbidden phrase error")
	}
}
