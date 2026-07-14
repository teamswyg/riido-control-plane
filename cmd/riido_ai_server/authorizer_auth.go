package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/authpep"
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func jwtAuthorizerFromEnv(domain riidoaiserver.RequestAuthorizer) (riidoaiserver.RequestAuthorizer, error) {
	issuer := strings.TrimSpace(os.Getenv(envAuthIssuer))
	resource := strings.TrimSpace(os.Getenv(envAuthResource))
	profile := strings.TrimSpace(os.Getenv(envAuthAuthorizationProfile))
	clientID := strings.TrimSpace(os.Getenv(envAuthIntrospectionClientID))
	secretRaw := os.Getenv(envAuthIntrospectionClientSecret)
	timeoutRaw := strings.TrimSpace(os.Getenv(envAuthHTTPTimeout))
	configured := issuer != "" || resource != "" || profile != "" || clientID != "" || secretRaw != "" || timeoutRaw != ""
	if !configured {
		return nil, nil
	}
	if issuer == "" || resource == "" || profile == "" || clientID == "" || len(secretRaw) < 32 || strings.TrimSpace(secretRaw) != secretRaw || strings.ContainsAny(secretRaw, "\r\n\x00") {
		return nil, fmt.Errorf("%s configuration requires exact issuer, resource, profile, client identity and a 32-byte secret", envAuthIssuer)
	}
	if domain == nil {
		return nil, fmt.Errorf("%s requires %s as the consumer-owned domain PDP", envAuthIssuer, envExternalAuthzURL)
	}
	timeout, err := envDurationSeconds(envAuthHTTPTimeout, 5*time.Second)
	if err != nil {
		return nil, err
	}
	if timeout > 10*time.Second {
		return nil, fmt.Errorf("%s must be at most 10 seconds", envAuthHTTPTimeout)
	}
	httpClient := &http.Client{Timeout: timeout}
	jwks, err := authpep.NewHTTPJWKSProvider(issuer, httpClient, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envAuthIssuer, err)
	}
	introspection, err := authpep.NewHTTPIntrospectionStatusResolver(httpClient, issuer, clientID, []byte(secretRaw), resource, profile, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envAuthIntrospectionClientID, err)
	}
	authorizer, err := authpep.NewJWTAccessTokenAuthorizer(authpep.JWTAuthorizationConfig{
		Issuer: issuer, Audience: resource, AuthorizationProfile: profile, ClockSkew: 30 * time.Second,
	}, jwks, introspection, domain)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envAuthIssuer, err)
	}
	return authorizer, nil
}
