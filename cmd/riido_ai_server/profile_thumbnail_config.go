package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func agentProfileThumbnailUploadServiceFromEnv() (riidoaiserver.AIAgentProfileThumbnailUploadService, error) {
	raw := profileThumbnailEnv()
	if raw.empty() {
		return nil, nil
	}
	if raw.bucket == "" {
		return nil, fmt.Errorf("%s is required when profile thumbnail upload is configured", envAgentProfileThumbnailBucket)
	}
	if raw.cdnBaseURL == "" {
		return nil, fmt.Errorf("%s is required when profile thumbnail upload is configured", envAgentProfileThumbnailCDNBase)
	}
	region := strings.TrimSpace(os.Getenv(envAWSRegion))
	if region == "" {
		return nil, fmt.Errorf("%s is required when profile thumbnail upload is configured", envAWSRegion)
	}
	maxBytes, err := envOptionalPositiveInt64(envAgentProfileThumbnailMaxBytes)
	if err != nil {
		return nil, err
	}
	expires, err := envOptionalDurationSeconds(envAgentProfileThumbnailExpires)
	if err != nil {
		return nil, err
	}
	provider, err := awsContainerCredentialsProviderFromEnvFor("profile thumbnail upload")
	if err != nil {
		return nil, err
	}
	service, err := riidoaiserver.NewS3AIAgentProfileThumbnailUploadService(riidoaiserver.S3AIAgentProfileThumbnailUploadConfig{
		Region:                region,
		Bucket:                raw.bucket,
		Prefix:                raw.prefix,
		CDNBaseURL:            raw.cdnBaseURL,
		UploadEndpoint:        raw.endpoint,
		MaxContentLengthBytes: maxBytes,
		Expires:               expires,
		CredentialsProvider:   provider,
	})
	if err != nil {
		return nil, fmt.Errorf("profile thumbnail upload: %w", err)
	}
	return service, nil
}
