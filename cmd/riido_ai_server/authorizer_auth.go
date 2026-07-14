package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	authclient "github.com/teamswyg/riido-auth-service/client"
	"github.com/teamswyg/riido-control-plane/internal/authpep"
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type authServiceStatusResolver struct {
	client *authclient.IntrospectionClient
}

func (r authServiceStatusResolver) ResolveAccessToken(ctx context.Context, raw string) (authpep.AccessTokenStatus, error) {
	if r.client == nil {
		return authpep.AccessTokenStatus{}, errors.New("auth introspection client is not configured")
	}
	result, err := r.client.Introspect(ctx, raw)
	if err != nil {
		return authpep.AccessTokenStatus{}, err
	}
	return authpep.AccessTokenStatus{
		Active: result.Active, Scope: result.Scope, ClientID: result.ClientID, Subject: result.Subject,
		TokenType: result.TokenType, ExpiresAt: result.ExpiresAt, IssuedAt: result.IssuedAt, NotBefore: result.NotBefore,
		Audience: result.Audience, Issuer: result.Issuer, JWTID: result.JWTID, Email: result.Email,
		EmailVerified: result.EmailVerified, AuthorizationProfile: result.AuthorizationProfile,
	}, nil
}

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
	introspection, err := authclient.NewIntrospectionClient(httpClient, issuer+"/oauth/introspect", clientID, []byte(secretRaw), resource, profile, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envAuthIntrospectionClientID, err)
	}
	authorizer, err := authpep.NewJWTAccessTokenAuthorizer(authpep.JWTAuthorizationConfig{
		Issuer: issuer, Audience: resource, AuthorizationProfile: profile, ClockSkew: 30 * time.Second,
	}, jwks, authServiceStatusResolver{client: introspection}, domain)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envAuthIssuer, err)
	}
	return authorizer, nil
}
