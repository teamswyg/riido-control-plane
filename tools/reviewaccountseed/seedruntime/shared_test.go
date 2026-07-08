package seedruntime

import "testing"

func TestReviewToken(t *testing.T) {
	if got := ReviewToken(); got != "review-token" {
		t.Fatalf("token = %q", got)
	}
}

func TestReviewProvisioningUsesHashedReviewToken(t *testing.T) {
	got, err := ReviewProvisioning()
	if err != nil {
		t.Fatalf("ReviewProvisioning: %v", err)
	}
	if got.Credential.Token != "" {
		t.Fatal("review provisioning must not expose raw token")
	}
	if got.Credential.TokenSHA256 != tokenHash(ReviewToken()) {
		t.Fatalf("token hash = %q", got.Credential.TokenSHA256)
	}
	if got.Credential.PrincipalID == "" || got.Principal.PrincipalID == "" {
		t.Fatalf("missing principal identity: %+v", got)
	}
	if len(got.AgentCatalogRecords) == 0 {
		t.Fatal("review provisioning should include catalog records")
	}
	if got.SyntheticProviderStatusAgentID == "" {
		t.Fatal("review provisioning should include synthetic provider status agent")
	}
}

func TestSameStrings(t *testing.T) {
	if !SameStrings([]string{"a", "b"}, []string{"a", "b"}) {
		t.Fatal("matching slices should compare equal")
	}
	if SameStrings([]string{"a"}, []string{"a", "b"}) {
		t.Fatal("different lengths should not compare equal")
	}
	if SameStrings([]string{"a", "b"}, []string{"b", "a"}) {
		t.Fatal("different order should not compare equal")
	}
}

func TestTokenHash(t *testing.T) {
	got := tokenHash("review-token")
	want := "3c9269e2a436bba87ad1255617f2231deeb2cb6a63f200e6cffb9c32585f8422"
	if got != want {
		t.Fatalf("hash = %q, want %q", got, want)
	}
}
