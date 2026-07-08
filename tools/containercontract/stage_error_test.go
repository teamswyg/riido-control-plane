package main

import (
	"strings"
	"testing"
)

func TestVerifyContractRejectsBuildStageDrift(t *testing.T) {
	for name, mutate := range map[string]func(string) string{
		"arg":      func(s string) string { return strings.Replace(s, "golang:1.26", "golang:1.25", 1) },
		"alias":    func(s string) string { return strings.Replace(s, " AS build", " AS builder", 1) },
		"workdir":  func(s string) string { return strings.Replace(s, "WORKDIR /src", "WORKDIR /app", 1) },
		"env":      func(s string) string { return strings.Replace(s, "CGO_ENABLED=0", "CGO_ENABLED=1", 1) },
		"download": func(s string) string { return strings.Replace(s, "go mod download", "go env", 1) },
		"go-build": func(s string) string { return strings.Replace(s, " -trimpath", "", 1) },
	} {
		t.Run(name, func(t *testing.T) {
			contract, _ := writeContainerContractFixture(t, "65532:65532", false)
			writeFixtureFile(t, contract.Dockerfile, mutate(fixtureDockerfile("65532:65532")))
			if _, err := verifyContract(contract); err == nil {
				t.Fatalf("expected build stage drift error")
			}
		})
	}
}

func TestVerifyContractRejectsFinalStageDrift(t *testing.T) {
	for name, mutate := range map[string]func(string) string{
		"base":       func(s string) string { return strings.Replace(s, "FROM scratch", "FROM alpine", 1) },
		"user":       func(s string) string { return strings.Replace(s, "USER 65532:65532", "USER 1000", 1) },
		"entrypoint": func(s string) string { return strings.Replace(s, "/riido_ai_server", "/other", 1) },
		"expose":     func(s string) string { return strings.Replace(s, "EXPOSE 8080", "EXPOSE 9090", 1) },
		"env":        func(s string) string { return strings.Replace(s, ":8080", ":9090", 1) },
		"copy":       func(s string) string { return strings.Replace(s, "/riido_ai_server", "/missing", 1) },
	} {
		t.Run(name, func(t *testing.T) {
			contract, _ := writeContainerContractFixture(t, "65532:65532", false)
			writeFixtureFile(t, contract.Dockerfile, mutate(fixtureDockerfile("65532:65532")))
			if _, err := verifyContract(contract); err == nil {
				t.Fatalf("expected final stage drift error")
			}
		})
	}
}
