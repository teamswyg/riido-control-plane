package main

import "testing"

func TestConfigFromEnvParsesAgentProfileThumbnailUpload(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAWSRegion, "ap-northeast-2")
	t.Setenv(envAWSContainerCredentialsFullURI, "http://169.254.170.2/credentials")
	t.Setenv(envAgentProfileThumbnailBucket, "profile-upload-test")
	t.Setenv(envAgentProfileThumbnailPrefix, "thumbnail/ai/profile/")
	t.Setenv(envAgentProfileThumbnailCDNBase, "https://cdn.example.test/")
	t.Setenv(envAgentProfileThumbnailMaxBytes, "1048576")
	t.Setenv(envAgentProfileThumbnailExpires, "60")

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.AIAgentProfileThumbnails == nil {
		t.Fatal("profile thumbnail upload service should be configured")
	}
}
