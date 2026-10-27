package riidoaiserver

import "testing"

func TestCLIInstallationResourceIsExactCanonicalHTTPSOrigin(t *testing.T) {
	credentials := []StaticTokenCredential{{PrincipalID: cliInstallationServicePrincipal, Token: "synthetic-service", Scopes: []string{"cli-installation:link"}}}
	for _, resource := range []string{"https://api.example.test/path", "https://api.example.test/", "https://api.example.test?", "https://api.example.test#", "https://api.example.test:0", "https://api.example.test:65536", "http://api.example.test", "https://user@api.example.test"} {
		if _, err := NewCLIInstallationServiceConfig(resource, credentials); err == nil {
			t.Fatalf("noncanonical installation resource accepted: %s", resource)
		}
	}
	for _, resource := range []string{"https://api.example.test", "https://api.example.test:443", "https://API.example.test", "https://[::1]:8443"} {
		if _, err := NewCLIInstallationServiceConfig(resource, credentials); err != nil {
			t.Fatalf("canonical resource refused: %s", resource)
		}
	}
}
