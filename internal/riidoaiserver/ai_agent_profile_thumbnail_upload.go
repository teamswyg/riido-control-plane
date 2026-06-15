package riidoaiserver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

const (
	defaultAgentProfileThumbnailPrefix   = "thumbnail/ai/profile/"
	defaultAgentProfileThumbnailMaxBytes = 5 * 1024 * 1024
	defaultAgentProfileThumbnailExpiry   = 5 * time.Minute
	agentProfileThumbnailUploadService   = "s3"
)

type AIAgentProfileThumbnailUploadService interface {
	CreateAIAgentProfileThumbnailUpload(ctx context.Context, principal AuthorizationResult, req CreateAgentProfileThumbnailUploadRequest) (AgentProfileThumbnailUploadResponse, error)
}

type S3AIAgentProfileThumbnailUploadConfig struct {
	Region                string
	Bucket                string
	Prefix                string
	CDNBaseURL            string
	UploadEndpoint        string
	MaxContentLengthBytes int64
	Expires               time.Duration
	CredentialsProvider   AWSCredentialsProvider
	Now                   func() time.Time
	Random                io.Reader
}

type S3AIAgentProfileThumbnailUploadService struct {
	region                string
	bucket                string
	prefix                string
	cdnBaseURL            string
	uploadURL             string
	maxContentLengthBytes int64
	expires               time.Duration
	credentialsProvider   AWSCredentialsProvider
	now                   func() time.Time
	random                io.Reader
}

func NewS3AIAgentProfileThumbnailUploadService(config S3AIAgentProfileThumbnailUploadConfig) (*S3AIAgentProfileThumbnailUploadService, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: profile thumbnail upload region is required")
	}
	bucket := strings.TrimSpace(config.Bucket)
	if bucket == "" {
		return nil, errors.New("riidoaiserver: profile thumbnail upload bucket is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: profile thumbnail upload credentials provider is required")
	}
	prefix := normalizeAgentProfileThumbnailPrefix(config.Prefix)
	cdnBaseURL, err := normalizeAgentProfileThumbnailCDNBaseURL(config.CDNBaseURL)
	if err != nil {
		return nil, err
	}
	uploadURL, err := normalizeAgentProfileThumbnailUploadURL(bucket, region, config.UploadEndpoint)
	if err != nil {
		return nil, err
	}
	maxBytes := config.MaxContentLengthBytes
	if maxBytes <= 0 {
		maxBytes = defaultAgentProfileThumbnailMaxBytes
	}
	expires := config.Expires
	if expires <= 0 {
		expires = defaultAgentProfileThumbnailExpiry
	}
	now := config.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	random := config.Random
	if random == nil {
		random = rand.Reader
	}
	return &S3AIAgentProfileThumbnailUploadService{
		region:                region,
		bucket:                bucket,
		prefix:                prefix,
		cdnBaseURL:            cdnBaseURL,
		uploadURL:             uploadURL,
		maxContentLengthBytes: maxBytes,
		expires:               expires,
		credentialsProvider:   config.CredentialsProvider,
		now:                   now,
		random:                random,
	}, nil
}

func (s *S3AIAgentProfileThumbnailUploadService) CreateAIAgentProfileThumbnailUpload(ctx context.Context, principal AuthorizationResult, req CreateAgentProfileThumbnailUploadRequest) (AgentProfileThumbnailUploadResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentProfileThumbnailUploadResponse{}, err
	}
	if s == nil {
		return AgentProfileThumbnailUploadResponse{}, errors.New("profile thumbnail upload service is not configured")
	}
	if strings.TrimSpace(principal.PrincipalID) == "" {
		return AgentProfileThumbnailUploadResponse{}, errors.New("principal_id is required")
	}
	contentType, ext, err := normalizeAgentProfileThumbnailContentType(req.ContentType)
	if err != nil {
		return AgentProfileThumbnailUploadResponse{}, err
	}
	if req.ContentLengthBytes <= 0 {
		return AgentProfileThumbnailUploadResponse{}, errors.New("content_length_bytes must be positive")
	}
	if req.ContentLengthBytes > s.maxContentLengthBytes {
		return AgentProfileThumbnailUploadResponse{}, fmt.Errorf("content_length_bytes must be %d bytes or fewer", s.maxContentLengthBytes)
	}
	credentials, err := cachedAWSCredentials(ctx, s.now, s.credentialsProvider, nil)
	if err != nil {
		return AgentProfileThumbnailUploadResponse{}, err
	}
	now := s.now().UTC()
	expiresAt := now.Add(s.expires)
	key, err := s.nextObjectKey(now, ext)
	if err != nil {
		return AgentProfileThumbnailUploadResponse{}, err
	}
	policy, err := s.postPolicy(credentials, key, contentType, now, expiresAt)
	if err != nil {
		return AgentProfileThumbnailUploadResponse{}, err
	}
	return AgentProfileThumbnailUploadResponse{
		SchemaVersion:         SchemaVersion,
		Method:                "POST",
		UploadURL:             s.uploadURL,
		FormFileField:         "file",
		FormFields:            policy.fields,
		ProfileThumbnailURL:   s.cdnBaseURL + "/" + key,
		ExpiresAt:             expiresAt,
		MaxContentLengthBytes: s.maxContentLengthBytes,
	}, nil
}

type agentProfileThumbnailPostPolicy struct {
	fields []AgentProfileThumbnailUploadFormField
}

func (s *S3AIAgentProfileThumbnailUploadService) postPolicy(credentials AWSCredentials, key, contentType string, now, expiresAt time.Time) (agentProfileThumbnailPostPolicy, error) {
	amzDate := now.UTC().Format("20060102T150405Z")
	dateScope := now.UTC().Format("20060102")
	credentialScope := strings.Join([]string{dateScope, s.region, agentProfileThumbnailUploadService, "aws4_request"}, "/")
	credential := credentials.AccessKeyID + "/" + credentialScope
	conditions := []any{
		map[string]string{"bucket": s.bucket},
		map[string]string{"key": key},
		map[string]string{"Content-Type": contentType},
		map[string]string{"success_action_status": "201"},
		[]any{"content-length-range", 1, s.maxContentLengthBytes},
		map[string]string{"x-amz-algorithm": "AWS4-HMAC-SHA256"},
		map[string]string{"x-amz-credential": credential},
		map[string]string{"x-amz-date": amzDate},
	}
	fields := []AgentProfileThumbnailUploadFormField{
		{Name: "key", Value: key},
		{Name: "Content-Type", Value: contentType},
		{Name: "success_action_status", Value: "201"},
		{Name: "x-amz-algorithm", Value: "AWS4-HMAC-SHA256"},
		{Name: "x-amz-credential", Value: credential},
		{Name: "x-amz-date", Value: amzDate},
	}
	if credentials.SessionToken != "" {
		conditions = append(conditions, map[string]string{"x-amz-security-token": credentials.SessionToken})
		fields = append(fields, AgentProfileThumbnailUploadFormField{Name: "x-amz-security-token", Value: credentials.SessionToken})
	}
	policyDocument := struct {
		Expiration string `json:"expiration"`
		Conditions []any  `json:"conditions"`
	}{
		Expiration: expiresAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Conditions: conditions,
	}
	policyJSON, err := json.Marshal(policyDocument)
	if err != nil {
		return agentProfileThumbnailPostPolicy{}, err
	}
	policy := base64.StdEncoding.EncodeToString(policyJSON)
	signingKey := awsV4SigningKey(credentials.SecretAccessKey, dateScope, s.region, agentProfileThumbnailUploadService)
	signature := hex.EncodeToString(hmacSHA256(signingKey, policy))
	fields = append(fields,
		AgentProfileThumbnailUploadFormField{Name: "policy", Value: policy},
		AgentProfileThumbnailUploadFormField{Name: "x-amz-signature", Value: signature},
	)
	return agentProfileThumbnailPostPolicy{fields: fields}, nil
}

func (s *S3AIAgentProfileThumbnailUploadService) nextObjectKey(now time.Time, ext string) (string, error) {
	var raw [16]byte
	if _, err := io.ReadFull(s.random, raw[:]); err != nil {
		return "", fmt.Errorf("profile thumbnail upload random id: %w", err)
	}
	return s.prefix + now.UTC().Format("20060102") + "-" + hex.EncodeToString(raw[:]) + ext, nil
}

func normalizeAgentProfileThumbnailPrefix(value string) string {
	value = strings.Trim(strings.TrimSpace(value), "/")
	if value == "" {
		value = strings.Trim(defaultAgentProfileThumbnailPrefix, "/")
	}
	return value + "/"
}

func normalizeAgentProfileThumbnailCDNBaseURL(value string) (string, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return "", errors.New("riidoaiserver: profile thumbnail CDN base URL is required")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse profile thumbnail CDN base URL: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		return "", errors.New("riidoaiserver: profile thumbnail CDN base URL must be an https URL")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("riidoaiserver: profile thumbnail CDN base URL must not include query or fragment")
	}
	return value, nil
}

func normalizeAgentProfileThumbnailUploadURL(bucket, region, endpoint string) (string, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", bucket, region)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parse profile thumbnail upload endpoint: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("riidoaiserver: profile thumbnail upload endpoint must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("riidoaiserver: profile thumbnail upload endpoint host is required")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("riidoaiserver: profile thumbnail upload endpoint must not include query or fragment")
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String(), nil
}

func normalizeAgentProfileThumbnailContentType(value string) (string, string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image/png":
		return "image/png", ".png", nil
	case "image/jpeg":
		return "image/jpeg", ".jpg", nil
	case "image/webp":
		return "image/webp", ".webp", nil
	default:
		return "", "", errors.New("content_type must be image/png, image/jpeg, or image/webp")
	}
}
