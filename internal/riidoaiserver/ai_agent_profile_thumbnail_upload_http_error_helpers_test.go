package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

const (
	profileThumbnailUploadV1Path = "/v1/client/ai-agent/profile-thumbnails/uploads"
	profileThumbnailUploadV2Path = "/v2/client/workspaces/ws-1/ai-agent/profile-thumbnails/uploads"
)

type thumbnailUploadStub struct {
	err error
}

func (s thumbnailUploadStub) CreateAIAgentProfileThumbnailUpload(
	context.Context,
	AuthorizationResult,
	CreateAgentProfileThumbnailUploadRequest,
) (AgentProfileThumbnailUploadResponse, error) {
	if s.err != nil {
		return AgentProfileThumbnailUploadResponse{}, s.err
	}
	return AgentProfileThumbnailUploadResponse{SchemaVersion: SchemaVersion}, nil
}

func thumbnailUploadTestHandler(t *testing.T, service AIAgentProfileThumbnailUploadService) http.Handler {
	t.Helper()
	return NewServer(ServerConfig{
		AIAgentProfileThumbnails: service,
		Authorizer: aiAgentClientHTTPAuthorizer(
			t,
			[]string{"ai-agent:write"},
			"user-1",
		),
	}).Handler()
}
