package main

import (
	"strings"
	"testing"
)

func TestGenerateReactQueryClientIncludesAIAgentSurface(t *testing.T) {
	spec := loadTestOpenAPI(t)
	got, err := generate(spec)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	body := string(got)
	assertGeneratedClientContainsAll(t, body, coreSurfaceRequired)
	if strings.Contains(body, "@tanstack/react-query") {
		t.Fatalf("core generated client must import React Query types through '@/lib/react-query'")
	}
	gotReact, err := generateReact(spec)
	if err != nil {
		t.Fatalf("generateReact: %v", err)
	}
	reactBody := string(gotReact)
	assertGeneratedClientContainsAll(t, reactBody, reactSurfaceRequired)
	if strings.Contains(reactBody, "core.Response") {
		t.Fatalf("generated react wrapper must use global Response, not core.Response")
	}
	if strings.Contains(reactBody, "@tanstack/react-query") {
		t.Fatalf("react generated client must import hooks through '@/lib/react-query'")
	}
}

func assertGeneratedClientContainsAll(t *testing.T, body string, required []string) {
	t.Helper()
	for _, item := range required {
		if !strings.Contains(body, item) {
			t.Fatalf("generated client missing %q", item)
		}
	}
}
