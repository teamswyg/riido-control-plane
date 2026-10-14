package deploypolicy

import "testing"

func assertNonCDRuntimeKeys(t *testing.T, keys []nonCDRuntimeKey) {
	t.Helper()
	for _, key := range expectedNonCDRuntimeKeys() {
		requireNonCDRuntimeKey(t, keys, key, "not a deploy/smoke GitHub configuration key")
	}
}

func expectedNonCDRuntimeKeys() []string {
	return []string{
		"RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT",
		"RIIDO_AI_SERVER_AI_AGENT_CLIENT_DYNAMODB_TABLE",
		"RIIDO_AI_SERVER_DYNAMODB_ENDPOINT",
		"RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_BUCKET",
		"RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_PREFIX",
		"RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_CDN_BASE_URL",
		"RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_MAX_BYTES",
		"RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_UPLOAD_EXPIRES_SECONDS",
		"RIIDO_AI_SERVER_AGENT_PROFILE_THUMBNAIL_S3_ENDPOINT",
		"RIIDO_AI_SERVER_ADDR",
	}
}

func requireNonCDRuntimeKey(t *testing.T, keys []nonCDRuntimeKey, wantName, wantReason string) {
	t.Helper()
	for _, key := range keys {
		if key.Name == wantName {
			requireContains(t, key.Reason, wantReason)
			return
		}
	}
	t.Fatalf("missing non-CD runtime key %q in %#v", wantName, keys)
}
