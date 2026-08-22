package controlplanegraphql

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/teamswyg/riido-control-plane/internal/controlplanegraphql/generated"
	"github.com/teamswyg/riido-control-plane/internal/controlplanehealth"
)

const BFFProductionSPIFFE = "spiffe://production.riido.io/service/riido-graphql-service"

func NewHandler(checker controlplanehealth.Checker) (http.Handler, error) {
	if checker == nil {
		return nil, fmt.Errorf("control-plane health use case is not configured")
	}
	if err := validatePublishedContract(); err != nil {
		return nil, err
	}
	server := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: &Resolver{Health: checker}}))
	server.AddTransport(transport.POST{})
	return requireWorkloadIdentity(server, BFFProductionSPIFFE), nil
}

func requireWorkloadIdentity(next http.Handler, expected string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !verifiedSPIFFE(r, expected) {
			http.Error(w, "workload_identity_required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func verifiedSPIFFE(r *http.Request, expected string) bool {
	return r != nil && r.TLS != nil && VerifyBFFProductionTLSConnection(*r.TLS) == nil && expected == BFFProductionSPIFFE
}

// VerifyBFFProductionTLSConnection binds the receiver to the one admitted BFF
// workload identity after the TLS stack has built a verified client chain.
func VerifyBFFProductionTLSConnection(state tls.ConnectionState) error {
	if len(state.VerifiedChains) == 0 || len(state.VerifiedChains[0]) == 0 {
		return fmt.Errorf("verified client certificate chain is required")
	}
	if !certificateHasURI(state.VerifiedChains[0][0], BFFProductionSPIFFE) {
		return fmt.Errorf("verified client certificate SPIFFE identity is not exact")
	}
	return nil
}

func certificateHasURI(certificate *x509.Certificate, expected string) bool {
	if certificate == nil || len(certificate.URIs) != 1 {
		return false
	}
	for _, identity := range certificate.URIs {
		if identity != nil && identity.String() == expected {
			return true
		}
	}
	return false
}
