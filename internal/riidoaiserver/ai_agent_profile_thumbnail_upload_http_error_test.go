package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIAgentClientProfileThumbnailUploadHTTPErrors(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		path    string
		body    string
		token   string
		service AIAgentProfileThumbnailUploadService
		want    int
	}{
		{
			name:   "method not allowed before service lookup",
			method: http.MethodGet,
			path:   profileThumbnailUploadV1Path,
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "service unavailable before auth",
			method: http.MethodPost,
			path:   profileThumbnailUploadV1Path,
			want:   http.StatusServiceUnavailable,
		},
		{
			name:    "unauthorized",
			method:  http.MethodPost,
			path:    profileThumbnailUploadV1Path,
			body:    `{}`,
			service: thumbnailUploadStub{},
			want:    http.StatusUnauthorized,
		},
		{
			name:    "malformed json",
			method:  http.MethodPost,
			path:    profileThumbnailUploadV2Path,
			body:    `{"content_type":`,
			token:   "ai-agent-token",
			service: thumbnailUploadStub{},
			want:    http.StatusBadRequest,
		},
		{
			name:    "service validation error",
			method:  http.MethodPost,
			path:    profileThumbnailUploadV2Path,
			body:    `{"content_type":"text/plain","content_length_bytes":1}`,
			token:   "ai-agent-token",
			service: thumbnailUploadStub{err: errors.New("content_type is not supported")},
			want:    http.StatusBadRequest,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := thumbnailUploadTestHandler(t, tc.service)
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("status=%d body=%s, want %d", resp.Code, resp.Body.String(), tc.want)
			}
		})
	}
}
