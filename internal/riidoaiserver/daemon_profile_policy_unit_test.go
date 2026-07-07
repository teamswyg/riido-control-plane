package riidoaiserver

import (
	"errors"
	"testing"
)

func TestNormalizeControlPlaneDaemonProfileAliases(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "", want: ""},
		{input: " development ", want: "development"},
		{input: "testnet", want: "staging"},
		{input: "STAGING", want: "staging"},
		{input: "Production", want: "production"},
	}
	for _, tt := range tests {
		got, err := normalizeControlPlaneDaemonProfile(tt.input)
		if err != nil {
			t.Fatalf("normalize %q: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("normalize %q = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDaemonProfileMatchingAndMismatchError(t *testing.T) {
	if !daemonProfileMatches("", "development") {
		t.Fatal("empty expected profile should accept any actual profile")
	}
	if !daemonProfileMatches("testnet", "staging") {
		t.Fatal("testnet expected profile should match staging actual profile")
	}
	if daemonProfileMatches("production", "staging") {
		t.Fatal("production profile should not match staging")
	}
	_, err := normalizeControlPlaneDaemonProfile("sandbox")
	if err == nil {
		t.Fatal("unsupported profile must return an error")
	}
	if !errors.Is(daemonProfileMismatchError("production", " staging "), ErrDaemonProfileMismatch) {
		t.Fatal("mismatch error must wrap ErrDaemonProfileMismatch")
	}
}
