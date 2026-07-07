package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func providerStatusAuthorizer(t *testing.T, scopes ...string) RequestAuthorizer {
	t.Helper()
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "daemon:agent-a",
		Token:       "agent-token",
		Scopes:      scopes,
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return auth
}

type providerStatusFailingReader struct{}

func (providerStatusFailingReader) GetProviderStatus(context.Context, string) (ProviderStatusSyncResponse, bool, error) {
	return ProviderStatusSyncResponse{}, false, errors.New("provider read failed")
}
