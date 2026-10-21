package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func assertAIAgentV2SmokeProfileUpload(t *testing.T, server http.Handler, base, token string) {
	t.Helper()
	profileUploadBytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/profile-thumbnails/uploads", token, `{"content_type":"image/jpeg","content_length_bytes":2048}`, http.StatusCreated)
	var profileUpload AgentProfileThumbnailUploadResponse
	aiAgentSmokeDecode(t, profileUploadBytes, &profileUpload)
	if profileUpload.Method != http.MethodPost ||
		profileUpload.FormFileField != "file" ||
		!strings.HasPrefix(profileUpload.ProfileThumbnailURL, "https://cdn.example.test/thumbnail/ai/profile/") ||
		!strings.HasSuffix(profileUpload.ProfileThumbnailURL, ".jpg") {
		t.Fatalf("v2 profile thumbnail upload intent = %+v", profileUpload)
	}
}
