package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func TestReviewAccountSeedProvisionsScopedCredentialWithoutRawSecret(t *testing.T) {
	seed, err := LoadReviewAccountSeed()
	if err != nil {
		t.Fatalf("LoadReviewAccountSeed: %v", err)
	}
	provisioning, err := ProvisionReviewAccount(seed, ReviewAccountProvisionInput{
		TokenSHA256: reviewTokenSHA256("review-token"),
	})
	if err != nil {
		t.Fatalf("ProvisionReviewAccount: %v", err)
	}
	if provisioning.Credential.Token != "" {
		t.Fatal("review account provisioning must not emit a raw token")
	}
	if provisioning.Credential.TokenSHA256 == "" {
		t.Fatal("review account provisioning should use token_sha256")
	}
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{provisioning.Credential})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	if _, err := authorizer.Authorize(context.Background(), "review-token", AuthorizationRequest{
		Resource: AuthorizationResourceMetrics,
		Action:   AuthorizationActionRead,
	}); err != nil {
		t.Fatalf("review account should read metrics: %v", err)
	}
	if _, err := authorizer.Authorize(context.Background(), "review-token", AuthorizationRequest{
		Resource: AuthorizationResourceAgentCatalog,
		Action:   AuthorizationActionRead,
	}); err != nil {
		t.Fatalf("review account should read agent catalog: %v", err)
	}
	if _, err := authorizer.Authorize(context.Background(), "review-token", AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionProviderStatusRead,
		AgentID:  provisioning.SyntheticProviderStatusAgentID,
	}); err != nil {
		t.Fatalf("review account should read synthetic provider status: %v", err)
	}
	if _, err := authorizer.Authorize(context.Background(), "review-token", AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionPoll,
		AgentID:  provisioning.SyntheticProviderStatusAgentID,
	}); err == nil {
		t.Fatal("review account must not poll as a daemon agent")
	}
}

func TestReviewAccountSeedExercisesCatalogVisibilityWithoutAdmin(t *testing.T) {
	seed, err := LoadReviewAccountSeed()
	if err != nil {
		t.Fatalf("LoadReviewAccountSeed: %v", err)
	}
	provisioning, err := ProvisionReviewAccount(seed, ReviewAccountProvisionInput{
		TokenSHA256: reviewTokenSHA256("review-token"),
	})
	if err != nil {
		t.Fatalf("ProvisionReviewAccount: %v", err)
	}
	if provisioning.Principal.HasRole(AgentCatalogRoleAdmin) {
		t.Fatal("review account should not be seeded as admin")
	}
	visible := VisibleAgentCatalogRecords(provisioning.Principal, provisioning.AgentCatalogRecords)
	if got, want := agentCatalogIDs(visible), []string{"review-owned-private", "review-owned-public", "review-other-public"}; !sameStrings(got, want) {
		t.Fatalf("review visible agents = %v, want %v", got, want)
	}
}

func TestReviewAccountSeedUsesOnlySyntheticNonRoutableProviderStatus(t *testing.T) {
	seed, err := LoadReviewAccountSeed()
	if err != nil {
		t.Fatalf("LoadReviewAccountSeed: %v", err)
	}
	provisioning, err := ProvisionReviewAccount(seed, ReviewAccountProvisionInput{
		TokenSHA256: reviewTokenSHA256("review-token"),
	})
	if err != nil {
		t.Fatalf("ProvisionReviewAccount: %v", err)
	}
	if provisioning.SyntheticProviderStatusRequest.DistributionChannel != hostintegration.DistributionChannelMacAppStore {
		t.Fatalf("unexpected review channel: %s", provisioning.SyntheticProviderStatusRequest.DistributionChannel)
	}
	for _, provider := range provisioning.SyntheticProviderStatusRequest.Providers {
		if provider.RoutingStatus == hostintegration.ProviderRoutingAvailable {
			t.Fatalf("review provider status must not be routable: %+v", provider)
		}
	}
}

func TestReviewAccountSeedArtifactContainsNoRawSecrets(t *testing.T) {
	seedBytes, err := reviewAccountSeedFS.ReadFile("review_account_seed.riido.json")
	if err != nil {
		t.Fatalf("read seed: %v", err)
	}
	if !json.Valid(seedBytes) {
		t.Fatal("seed is not valid json")
	}
	encoded := string(seedBytes)
	for _, forbidden := range []string{"\"token\"", "\"password\"", "provider_executable_path", "workspace_root_path", "api_key"} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("review account seed leaked forbidden field %s: %s", forbidden, encoded)
		}
	}
}

func reviewTokenSHA256(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
