package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/controlplanegraphql"
)

const controlPlaneGraphQLServerName = "control-plane.services.production.riido.internal"

func controlPlaneGraphQLMTLSConfigFromEnv() (controlPlaneGraphQLMTLSConfig, error) {
	addr := strings.TrimSpace(os.Getenv(envControlPlaneGraphQLMTLSAddr))
	certFile := strings.TrimSpace(os.Getenv(envControlPlaneGraphQLServerCertFile))
	keyFile := strings.TrimSpace(os.Getenv(envControlPlaneGraphQLServerKeyFile))
	caFile := strings.TrimSpace(os.Getenv(envControlPlaneGraphQLClientCAFile))
	configured := 0
	for _, value := range []string{addr, certFile, keyFile, caFile} {
		if value != "" {
			configured++
		}
	}
	if configured == 0 {
		return controlPlaneGraphQLMTLSConfig{}, nil
	}
	if configured != 4 {
		return controlPlaneGraphQLMTLSConfig{}, fmt.Errorf("control-plane GraphQL mTLS requires address, server certificate, server key, and client CA")
	}

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return controlPlaneGraphQLMTLSConfig{}, fmt.Errorf("load control-plane GraphQL server identity: %w", err)
	}
	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		return controlPlaneGraphQLMTLSConfig{}, fmt.Errorf("parse control-plane GraphQL server identity: %w", err)
	}
	if err := validateControlPlaneGraphQLServerLeaf(leaf, time.Now()); err != nil {
		return controlPlaneGraphQLMTLSConfig{}, err
	}
	certificate.Leaf = leaf

	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return controlPlaneGraphQLMTLSConfig{}, fmt.Errorf("read control-plane GraphQL client CA: %w", err)
	}
	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(caPEM) {
		return controlPlaneGraphQLMTLSConfig{}, fmt.Errorf("control-plane GraphQL client CA contains no certificates")
	}
	tlsConfig := &tls.Config{
		Certificates:     []tls.Certificate{certificate},
		ClientAuth:       tls.RequireAndVerifyClientCert,
		ClientCAs:        clientCAs,
		MinVersion:       tls.VersionTLS13,
		MaxVersion:       tls.VersionTLS13,
		VerifyConnection: controlplanegraphql.VerifyBFFProductionTLSConnection,
	}
	return controlPlaneGraphQLMTLSConfig{Addr: addr, TLSConfig: tlsConfig}, nil
}

func validateControlPlaneGraphQLServerLeaf(leaf *x509.Certificate, now time.Time) error {
	if leaf == nil || leaf.IsCA || now.Before(leaf.NotBefore) || now.After(leaf.NotAfter) {
		return fmt.Errorf("control-plane GraphQL server certificate is not a current leaf")
	}
	if leaf.KeyUsage&x509.KeyUsageDigitalSignature == 0 || len(leaf.ExtKeyUsage) != 1 || leaf.ExtKeyUsage[0] != x509.ExtKeyUsageServerAuth {
		return fmt.Errorf("control-plane GraphQL server certificate usage is not exact")
	}
	if len(leaf.DNSNames) != 1 || leaf.DNSNames[0] != controlPlaneGraphQLServerName || leaf.VerifyHostname(controlPlaneGraphQLServerName) != nil {
		return fmt.Errorf("control-plane GraphQL server certificate DNS identity is not exact")
	}
	return nil
}
