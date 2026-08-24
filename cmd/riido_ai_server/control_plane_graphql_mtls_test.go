package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/controlplanegraphql"
)

func TestControlPlaneGraphQLDedicatedMTLSListener(t *testing.T) {
	now := time.Now()
	trusted := newTestCertificateAuthority(t, "trusted")
	other := newTestCertificateAuthority(t, "other")
	serverCertificate := trusted.issue(t, certificateSpec{
		commonName: controlPlaneGraphQLServerName,
		dnsNames:   []string{controlPlaneGraphQLServerName},
		extUsage:   []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		notBefore:  now.Add(-time.Hour),
		notAfter:   now.Add(time.Hour),
	})
	config := loadTestControlPlaneGraphQLMTLSConfig(t, trusted, serverCertificate)
	config.Addr = availableTestAddress(t)
	servers, err := newRuntimeHTTPServers(runtimeConfig{
		Addr:                    "127.0.0.1:0",
		ControlPlaneGraphQLMTLS: config,
	}, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 2 || servers[1].TLSConfig == nil {
		t.Fatalf("dedicated mTLS server missing: %#v", servers)
	}
	servers[1].ErrorLog = discardTestLogger()
	errCh := listenAndServeAll([]*http.Server{servers[1]})
	defer func() {
		_ = servers[1].Shutdown(context.Background())
		if err := <-errCh; err != nil {
			t.Errorf("mTLS server shutdown: %v", err)
		}
	}()
	endpoint := "https://" + config.Addr

	valid := trusted.issue(t, certificateSpec{
		commonName: "riido-graphql-service",
		uris:       []string{controlplanegraphql.BFFProductionSPIFFE},
		extUsage:   []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		notBefore:  now.Add(-time.Hour),
		notAfter:   now.Add(time.Hour),
	})
	status, body, err := callOwnerHealthEventually(endpoint, trusted, valid)
	if err != nil || status != http.StatusOK || !strings.Contains(body, `"healthCheck":200`) {
		t.Fatalf("valid mTLS call status=%d body=%s err=%v", status, body, err)
	}

	tests := []struct {
		name       string
		issuer     *testCertificateAuthority
		spec       certificateSpec
		maxVersion uint16
	}{
		{name: "wrong spiffe", issuer: trusted, spec: validClientSpec(now, "spiffe://production.riido.io/service/other")},
		{name: "multiple spiffe uris", issuer: trusted, spec: certificateSpec{commonName: "bff", uris: []string{controlplanegraphql.BFFProductionSPIFFE, "spiffe://production.riido.io/service/other"}, extUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, notBefore: now.Add(-time.Hour), notAfter: now.Add(time.Hour)}},
		{name: "wrong eku", issuer: trusted, spec: certificateSpec{commonName: "bff", uris: []string{controlplanegraphql.BFFProductionSPIFFE}, extUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, notBefore: now.Add(-time.Hour), notAfter: now.Add(time.Hour)}},
		{name: "expired", issuer: trusted, spec: certificateSpec{commonName: "bff", uris: []string{controlplanegraphql.BFFProductionSPIFFE}, extUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, notBefore: now.Add(-2 * time.Hour), notAfter: now.Add(-time.Hour)}},
		{name: "not yet valid", issuer: trusted, spec: certificateSpec{commonName: "bff", uris: []string{controlplanegraphql.BFFProductionSPIFFE}, extUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, notBefore: now.Add(time.Hour), notAfter: now.Add(2 * time.Hour)}},
		{name: "wrong ca", issuer: other, spec: validClientSpec(now, controlplanegraphql.BFFProductionSPIFFE)},
		{name: "tls below 1.3", issuer: trusted, spec: validClientSpec(now, controlplanegraphql.BFFProductionSPIFFE), maxVersion: tls.VersionTLS12},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			certificate := test.issuer.issue(t, test.spec)
			if _, _, err := callOwnerHealth(endpoint, trusted, certificate, test.maxVersion); err == nil {
				t.Fatal("unadmitted transport reached the owner")
			}
		})
	}
}

func callOwnerHealthEventually(endpoint string, roots *testCertificateAuthority, certificate tls.Certificate) (int, string, error) {
	var status int
	var body string
	var err error
	for attempt := 0; attempt < 50; attempt++ {
		status, body, err = callOwnerHealth(endpoint, roots, certificate, tls.VersionTLS13)
		if err == nil {
			return status, body, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return status, body, err
}

func availableTestAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}

func TestControlPlaneGraphQLMTLSConfigurationFailsClosed(t *testing.T) {
	clearRiidoAIServerEnv(t)
	if config, err := controlPlaneGraphQLMTLSConfigFromEnv(); err != nil || config.TLSConfig != nil {
		t.Fatalf("disabled config=%#v err=%v", config, err)
	}
	t.Setenv(envControlPlaneGraphQLMTLSAddr, "127.0.0.1:8443")
	if _, err := controlPlaneGraphQLMTLSConfigFromEnv(); err == nil {
		t.Fatal("partial mTLS configuration was admitted")
	}
	if err := validateControlPlaneGraphQLMTLSRuntimeConfig(runtimeConfig{
		Addr: "127.0.0.1:8080",
		ControlPlaneGraphQLMTLS: controlPlaneGraphQLMTLSConfig{
			Addr: "127.0.0.1:8443", TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}); err == nil {
		t.Fatal("generic TLS configuration was admitted")
	}
}

func validClientSpec(now time.Time, identity string) certificateSpec {
	return certificateSpec{commonName: "bff", uris: []string{identity}, extUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, notBefore: now.Add(-time.Hour), notAfter: now.Add(time.Hour)}
}

func callOwnerHealth(endpoint string, roots *testCertificateAuthority, certificate tls.Certificate, maxVersion uint16) (int, string, error) {
	if maxVersion == 0 {
		maxVersion = tls.VersionTLS13
	}
	rootPool := x509.NewCertPool()
	rootPool.AppendCertsFromPEM(roots.pem)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{
		Certificates: []tls.Certificate{certificate}, RootCAs: rootPool,
		ServerName: controlPlaneGraphQLServerName, MinVersion: tls.VersionTLS12, MaxVersion: maxVersion,
	}}}
	payload, _ := json.Marshal(map[string]string{"query": "query OwnerHealth { healthCheck }"})
	request, _ := http.NewRequest(http.MethodPost, endpoint+"/graphql", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return 0, "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	return response.StatusCode, string(body), err
}

func loadTestControlPlaneGraphQLMTLSConfig(t *testing.T, authority *testCertificateAuthority, certificate tls.Certificate) controlPlaneGraphQLMTLSConfig {
	t.Helper()
	directory := t.TempDir()
	certFile := filepath.Join(directory, "server.crt")
	keyFile := filepath.Join(directory, "server.key")
	caFile := filepath.Join(directory, "client-ca.crt")
	writeTestFile(t, certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Certificate[0]}))
	key, _ := x509.MarshalECPrivateKey(certificate.PrivateKey.(*ecdsa.PrivateKey))
	writeTestFile(t, keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: key}))
	writeTestFile(t, caFile, authority.pem)
	t.Setenv(envControlPlaneGraphQLMTLSAddr, "127.0.0.1:8443")
	t.Setenv(envControlPlaneGraphQLServerCertFile, certFile)
	t.Setenv(envControlPlaneGraphQLServerKeyFile, keyFile)
	t.Setenv(envControlPlaneGraphQLClientCAFile, caFile)
	config, err := controlPlaneGraphQLMTLSConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	return config
}

func writeTestFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func discardTestLogger() *log.Logger { return log.New(io.Discard, "", 0) }

type testCertificateAuthority struct {
	certificate *x509.Certificate
	key         *ecdsa.PrivateKey
	pem         []byte
}

type certificateSpec struct {
	commonName string
	dnsNames   []string
	uris       []string
	extUsage   []x509.ExtKeyUsage
	notBefore  time.Time
	notAfter   time.Time
}

func newTestCertificateAuthority(t *testing.T, name string) *testCertificateAuthority {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	template := &x509.Certificate{SerialNumber: big.NewInt(time.Now().UnixNano()), Subject: pkix.Name{CommonName: name}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certificate, _ := x509.ParseCertificate(der)
	return &testCertificateAuthority{certificate: certificate, key: key, pem: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})}
}

func (authority *testCertificateAuthority) issue(t *testing.T, spec certificateSpec) tls.Certificate {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	template := &x509.Certificate{SerialNumber: big.NewInt(time.Now().UnixNano()), Subject: pkix.Name{CommonName: spec.commonName}, DNSNames: spec.dnsNames, NotBefore: spec.notBefore, NotAfter: spec.notAfter, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: spec.extUsage}
	for _, rawURI := range spec.uris {
		identity, err := url.Parse(rawURI)
		if err != nil {
			t.Fatal(err)
		}
		template.URIs = append(template.URIs, identity)
	}
	der, err := x509.CreateCertificate(rand.Reader, template, authority.certificate, &key.PublicKey, authority.key)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der, authority.certificate.Raw}, PrivateKey: key}
}
