package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestS3AIAgentProfileThumbnailUploadCreatesPostPolicy(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKIDEXAMPLE", "SECRET", "SESSION")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	now := time.Date(2026, 6, 15, 9, 30, 0, 0, time.UTC)
	service, err := NewS3AIAgentProfileThumbnailUploadService(S3AIAgentProfileThumbnailUploadConfig{
		Region:                "ap-northeast-2",
		Bucket:                "profile-upload-test",
		Prefix:                "thumbnail/ai/profile/",
		CDNBaseURL:            "https://cdn.example.test/",
		MaxContentLengthBytes: 1024 * 1024,
		Expires:               time.Minute,
		CredentialsProvider:   provider,
		Now:                   func() time.Time { return now },
		Random:                bytes.NewReader([]byte("0123456789abcdef")),
	})
	if err != nil {
		t.Fatalf("NewS3AIAgentProfileThumbnailUploadService: %v", err)
	}

	resp, err := service.CreateAIAgentProfileThumbnailUpload(context.Background(), AuthorizationResult{PrincipalID: "user-1"}, CreateAgentProfileThumbnailUploadRequest{
		ContentType:        "image/png",
		ContentLengthBytes: 1234,
		FileName:           "avatar.png",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentProfileThumbnailUpload: %v", err)
	}
	if resp.Method != http.MethodPost || resp.FormFileField != "file" {
		t.Fatalf("upload method/file field = %s/%s", resp.Method, resp.FormFileField)
	}
	if resp.UploadURL != "https://profile-upload-test.s3.ap-northeast-2.amazonaws.com/" {
		t.Fatalf("upload_url = %q", resp.UploadURL)
	}
	if want := "https://cdn.example.test/thumbnail/ai/profile/20260615-30313233343536373839616263646566.png"; resp.ProfileThumbnailURL != want {
		t.Fatalf("profile_thumbnail_url = %q, want %q", resp.ProfileThumbnailURL, want)
	}
	fields := profileUploadFields(resp.FormFields)
	if fields["Content-Type"] != "image/png" ||
		fields["x-amz-security-token"] != "SESSION" ||
		fields["success_action_status"] != "201" {
		t.Fatalf("unexpected form fields: %#v", fields)
	}
	policyBody, err := base64.StdEncoding.DecodeString(fields["policy"])
	if err != nil {
		t.Fatalf("decode policy: %v", err)
	}
	policyText := string(policyBody)
	for _, needle := range []string{
		`"bucket":"profile-upload-test"`,
		`"key":"thumbnail/ai/profile/20260615-30313233343536373839616263646566.png"`,
		`"Content-Type":"image/png"`,
		`"content-length-range",1,1048576`,
	} {
		if !strings.Contains(policyText, needle) {
			t.Fatalf("policy %q missing %q", policyText, needle)
		}
	}
	if strings.Contains(policyText, "SECRET") {
		t.Fatalf("policy must not expose AWS secret key: %q", policyText)
	}
}

func TestAIAgentClientProfileThumbnailUploadEndpoint(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKIDEXAMPLE", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	service, err := NewS3AIAgentProfileThumbnailUploadService(S3AIAgentProfileThumbnailUploadConfig{
		Region:              "ap-northeast-2",
		Bucket:              "profile-upload-test",
		CDNBaseURL:          "https://cdn.example.test",
		CredentialsProvider: provider,
		Now:                 func() time.Time { return time.Date(2026, 6, 15, 9, 30, 0, 0, time.UTC) },
		Random:              bytes.NewReader([]byte("abcdefghijklmnop")),
	})
	if err != nil {
		t.Fatalf("NewS3AIAgentProfileThumbnailUploadService: %v", err)
	}
	authorizer := aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:write"}, "user-1")
	handler := NewServer(ServerConfig{
		AIAgentProfileThumbnails: service,
		Authorizer:               authorizer,
	}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/profile-thumbnails/uploads", strings.NewReader(`{"content_type":"image/webp","content_length_bytes":2048}`))
	req.Header.Set("Authorization", "Bearer ai-agent-token")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AgentProfileThumbnailUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.HasPrefix(out.ProfileThumbnailURL, "https://cdn.example.test/thumbnail/ai/profile/") ||
		!strings.HasSuffix(out.ProfileThumbnailURL, ".webp") {
		t.Fatalf("profile_thumbnail_url = %q", out.ProfileThumbnailURL)
	}
}

func profileUploadFields(fields []AgentProfileThumbnailUploadFormField) map[string]string {
	out := make(map[string]string, len(fields))
	for _, field := range fields {
		out[field.Name] = field.Value
	}
	return out
}
