package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func newAIAgentClientHTTPTestServer(t *testing.T, credentials []StaticTokenCredential) http.Handler {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer(credentials)
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() {
		assignmentStore.Close()
	})
	profileThumbnailCredentials, err := NewStaticAWSCredentialsProvider("AKIDEXAMPLE", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	profileThumbnails, err := NewS3AIAgentProfileThumbnailUploadService(S3AIAgentProfileThumbnailUploadConfig{
		Region:              "ap-northeast-2",
		Bucket:              "profile-upload-test",
		CDNBaseURL:          "https://cdn.example.test",
		CredentialsProvider: profileThumbnailCredentials,
		Now:                 func() time.Time { return time.Date(2026, 6, 15, 9, 30, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewS3AIAgentProfileThumbnailUploadService: %v", err)
	}
	return NewServer(ServerConfig{
		AIAgentClient:            aiAgentStore,
		AIAgentProfileThumbnails: profileThumbnails,
		Assignment:               assignmentStore,
		TaskContext:              &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:               authorizer,
	}).Handler()
}

func aiAgentClientHTTPAuthorizer(t *testing.T, scopes []string, principalID string) RequestAuthorizer {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: principalID,
		Token:       "ai-agent-token",
		Scopes:      scopes,
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return authorizer
}
