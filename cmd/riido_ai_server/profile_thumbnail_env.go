package main

import (
	"os"
	"strings"
)

type profileThumbnailRawEnv struct {
	bucket      string
	prefix      string
	cdnBaseURL  string
	maxBytesRaw string
	expiresRaw  string
	endpoint    string
}

func profileThumbnailEnv() profileThumbnailRawEnv {
	return profileThumbnailRawEnv{
		bucket:      strings.TrimSpace(os.Getenv(envAgentProfileThumbnailBucket)),
		prefix:      strings.TrimSpace(os.Getenv(envAgentProfileThumbnailPrefix)),
		cdnBaseURL:  strings.TrimSpace(os.Getenv(envAgentProfileThumbnailCDNBase)),
		maxBytesRaw: strings.TrimSpace(os.Getenv(envAgentProfileThumbnailMaxBytes)),
		expiresRaw:  strings.TrimSpace(os.Getenv(envAgentProfileThumbnailExpires)),
		endpoint:    strings.TrimSpace(os.Getenv(envAgentProfileThumbnailS3Endpoint)),
	}
}

func (raw profileThumbnailRawEnv) empty() bool {
	return raw.bucket == "" &&
		raw.prefix == "" &&
		raw.cdnBaseURL == "" &&
		raw.maxBytesRaw == "" &&
		raw.expiresRaw == "" &&
		raw.endpoint == ""
}
