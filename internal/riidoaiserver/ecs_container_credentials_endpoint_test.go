package riidoaiserver

import "testing"

func TestECSContainerCredentialsProviderRejectsUnsafeEndpoint(t *testing.T) {
	for _, endpoint := range []string{
		"ftp://169.254.170.2/credentials",
		"http://metadata.example.com/credentials",
		"http://192.0.2.10/credentials",
		"https://user:pass@metadata.example.com/credentials",
		"https://metadata.example.com/credentials?token=raw",
		"https://metadata.example.com/credentials#fragment",
	} {
		t.Run(endpoint, func(t *testing.T) {
			if _, err := NewECSContainerCredentialsProvider(ECSContainerCredentialsProviderConfig{Endpoint: endpoint}); err == nil {
				t.Fatalf("expected endpoint %q to be rejected", endpoint)
			}
		})
	}
}

func TestECSContainerCredentialsProviderAllowsPlainHTTPOnlyForMetadataAddresses(t *testing.T) {
	for _, endpoint := range []string{
		"http://localhost/credentials",
		"http://127.0.0.1:8080/credentials",
		"http://[::1]:8080/credentials",
		"http://169.254.170.2/credentials",
		"http://169.254.170.23/credentials",
	} {
		t.Run(endpoint, func(t *testing.T) {
			if _, err := NewECSContainerCredentialsProvider(ECSContainerCredentialsProviderConfig{Endpoint: endpoint}); err != nil {
				t.Fatalf("expected metadata endpoint %q to be allowed: %v", endpoint, err)
			}
		})
	}
}
