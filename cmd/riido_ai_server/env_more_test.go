package main

import (
	"strings"
	"testing"
)

func TestEnvOptionalBoolParsesAcceptedValues(t *testing.T) {
	for _, tt := range []struct {
		value string
		want  bool
	}{
		{" YES ", true},
		{"on", true},
		{"0", false},
		{"Off", false},
	} {
		t.Run(tt.value, func(t *testing.T) {
			t.Setenv(envAIAgentClientDev, tt.value)
			got, err := envOptionalBool(envAIAgentClientDev)
			if err != nil || got != tt.want {
				t.Fatalf("envOptionalBool(%q) = %v, %v; want %v", tt.value, got, err, tt.want)
			}
		})
	}
}

func TestEnvOptionalBoolRejectsUnknownValue(t *testing.T) {
	t.Setenv(envAIAgentClientDev, "maybe")
	if _, err := envOptionalBool(envAIAgentClientDev); err == nil ||
		!strings.Contains(err.Error(), envAIAgentClientDev) {
		t.Fatalf("envOptionalBool err=%v", err)
	}
}

func TestEnvOptionalPositiveInt64ParsesAndRejectsValues(t *testing.T) {
	t.Setenv(envAgentProfileThumbnailMaxBytes, "42")
	got, err := envOptionalPositiveInt64(envAgentProfileThumbnailMaxBytes)
	if err != nil || got != 42 {
		t.Fatalf("envOptionalPositiveInt64 = %d, %v; want 42", got, err)
	}
	for _, value := range []string{"0", "-1", "nope"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(envAgentProfileThumbnailMaxBytes, value)
			_, err := envOptionalPositiveInt64(envAgentProfileThumbnailMaxBytes)
			if err == nil || !strings.Contains(err.Error(), envAgentProfileThumbnailMaxBytes) {
				t.Fatalf("envOptionalPositiveInt64 err=%v", err)
			}
		})
	}
}
