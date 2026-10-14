package main

import (
	"strings"
	"testing"
)

func TestProfileThumbnailUploadNoopsWhenUnset(t *testing.T) {
	clearRiidoAIServerEnv(t)
	service, err := agentProfileThumbnailUploadServiceFromEnv()
	if err != nil || service != nil {
		t.Fatalf("agentProfileThumbnailUploadServiceFromEnv = %T, %v; want nil, nil", service, err)
	}
}

func TestProfileThumbnailUploadRejectsPartialConfig(t *testing.T) {
	for _, tt := range []struct {
		name string
		set  func(*testing.T)
		want string
	}{
		{"missing bucket", func(t *testing.T) {
			t.Helper()
			t.Setenv(envAgentProfileThumbnailPrefix, "x")
		}, envAgentProfileThumbnailBucket},
		{"missing cdn", func(t *testing.T) {
			t.Helper()
			t.Setenv(envAgentProfileThumbnailBucket, "bucket")
		}, envAgentProfileThumbnailCDNBase},
		{"missing region", func(t *testing.T) {
			t.Helper()
			t.Setenv(envAgentProfileThumbnailBucket, "bucket")
			t.Setenv(envAgentProfileThumbnailCDNBase, "https://cdn.example.test")
		}, envAWSRegion},
	} {
		t.Run(tt.name, func(t *testing.T) {
			clearRiidoAIServerEnv(t)
			tt.set(t)
			_, err := agentProfileThumbnailUploadServiceFromEnv()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("agentProfileThumbnailUploadServiceFromEnv err=%v, want %s", err, tt.want)
			}
		})
	}
}

func TestProfileThumbnailRawEnvDetectsAnyConfiguredField(t *testing.T) {
	raw := profileThumbnailRawEnv{endpoint: "http://127.0.0.1:9000"}
	if raw.empty() {
		t.Fatal("profile thumbnail env with endpoint should not be empty")
	}
}
