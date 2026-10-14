package authpep

import (
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestIssuerAuthorizationSchemeSelectorOwnsConfiguredIssuer(t *testing.T) {
	selector, err := NewIssuerAuthorizationSchemeSelector("https://auth.riido.io")
	if err != nil {
		t.Fatal(err)
	}
	for name, token := range map[string]string{
		"exact issuer":     jwtShapedTokenForTest(t, `{"iss":"https://auth.riido.io"}`),
		"duplicate issuer": jwtShapedTokenForTest(t, `{"iss":"legacy","iss":"https://auth.riido.io"}`),
	} {
		t.Run(name, func(t *testing.T) {
			scheme, selectErr := selector.SelectAuthorizationScheme(token)
			if selectErr != nil || scheme != riidoaiserver.AuthorizationSchemeAuthServiceV2 {
				t.Fatalf("scheme=%q err=%v", scheme, selectErr)
			}
		})
	}
}

func TestIssuerAuthorizationSchemeSelectorLeavesOtherCredentialsToLegacy(t *testing.T) {
	selector, _ := NewIssuerAuthorizationSchemeSelector("https://auth.riido.io")
	for name, token := range map[string]string{
		"opaque":  "legacy-token",
		"missing": jwtShapedTokenForTest(t, `{"sub":"legacy"}`),
		"foreign": jwtShapedTokenForTest(t, `{"iss":"https://legacy.example"}`),
	} {
		t.Run(name, func(t *testing.T) {
			scheme, err := selector.SelectAuthorizationScheme(token)
			if err != nil || scheme != riidoaiserver.AuthorizationSchemeLegacyV1 {
				t.Fatalf("scheme=%q err=%v", scheme, err)
			}
		})
	}
}

func TestIssuerAuthorizationSchemeSelectorRejectsNonHTTPSIssuer(t *testing.T) {
	if _, err := NewIssuerAuthorizationSchemeSelector("http://auth.riido.io"); err == nil {
		t.Fatal("expected non-HTTPS issuer rejection")
	}
}
