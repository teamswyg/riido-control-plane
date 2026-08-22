package controlplanegraphql

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/controlplanehealth"
)

type checkerFunc func(context.Context) (int, error)

func (function checkerFunc) HealthCheck(ctx context.Context) (int, error) { return function(ctx) }

func TestHealthCheckGeneratedReceiverReturnsExactFrozenValueWithoutUserAuth(t *testing.T) {
	handler, err := NewHandler(controlplanehealth.NewService())
	if err != nil {
		t.Fatal(err)
	}
	response := performGraphQL(t, handler, `query OwnerHealth { healthCheck }`, BFFProductionSPIFFE)
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status=%d headers=%v body=%s", response.Code, response.Header(), response.Body.String())
	}
	var payload struct {
		Data   map[string]int `json:"data"`
		Errors []any          `json:"errors"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Errors) != 0 || payload.Data["healthCheck"] != 200 {
		t.Fatalf("unexpected owner response %s", response.Body.String())
	}
}

func TestGeneratedReceiverRejectsNonExactNullAndOwnerErrors(t *testing.T) {
	tests := []struct {
		name    string
		checker checkerFunc
	}{
		{name: "wrong integer", checker: func(context.Context) (int, error) { return 201, nil }},
		{name: "owner error", checker: func(context.Context) (int, error) { return 0, errors.New("unhealthy") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewHandler(test.checker)
			if err != nil {
				t.Fatal(err)
			}
			response := performGraphQL(t, handler, `query OwnerHealth { healthCheck }`, BFFProductionSPIFFE)
			if response.Code != http.StatusOK || bytes.Contains(response.Body.Bytes(), []byte(`"healthCheck":200`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"errors"`)) {
				t.Fatalf("invalid owner result escaped: %s", response.Body.String())
			}
		})
	}
}

func TestFireErrorAlwaysReturnsGraphQLErrorWithoutPanicking(t *testing.T) {
	handler, err := NewHandler(controlplanehealth.NewService())
	if err != nil {
		t.Fatal(err)
	}
	response := performGraphQL(t, handler, `query OwnerHealth { fireError }`, BFFProductionSPIFFE)
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"fireError":null`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"errors"`)) {
		t.Fatalf("fireError did not fail safely: %s", response.Body.String())
	}
}

func TestReceiverRequiresExactBFFWorkloadIdentity(t *testing.T) {
	handler, err := NewHandler(controlplanehealth.NewService())
	if err != nil {
		t.Fatal(err)
	}
	for _, identity := range []string{"", "spiffe://production.riido.io/service/other"} {
		response := performGraphQL(t, handler, `query OwnerHealth { healthCheck }`, identity)
		if response.Code != http.StatusUnauthorized || bytes.Contains(response.Body.Bytes(), []byte("200")) {
			t.Fatalf("identity %q status=%d body=%s", identity, response.Code, response.Body.String())
		}
	}
	response := performGraphQLWithIdentities(t, handler, `query OwnerHealth { healthCheck }`, []string{BFFProductionSPIFFE, "spiffe://production.riido.io/service/other"})
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("multiple URI identities were admitted: %d %s", response.Code, response.Body.String())
	}
}

func TestHealthReceiverIsStatelessAcrossRestartAndMultipleInstances(t *testing.T) {
	for instance := 0; instance < 2; instance++ {
		handler, err := NewHandler(controlplanehealth.NewService())
		if err != nil {
			t.Fatal(err)
		}
		for request := 0; request < 2; request++ {
			response := performGraphQL(t, handler, `query OwnerHealth { healthCheck }`, BFFProductionSPIFFE)
			if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"healthCheck":200`)) {
				t.Fatalf("instance=%d request=%d body=%s", instance, request, response.Body.String())
			}
		}
	}
}

func performGraphQL(t *testing.T, handler http.Handler, query, identity string) *httptest.ResponseRecorder {
	t.Helper()
	identities := []string{}
	if identity != "" {
		identities = append(identities, identity)
	}
	return performGraphQLWithIdentities(t, handler, query, identities)
}

func performGraphQLWithIdentities(t *testing.T, handler http.Handler, query string, identities []string) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"query": query, "operationName": "OwnerHealth"})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(raw))
	request.Header.Set("Content-Type", "application/json")
	if len(identities) != 0 {
		uris := make([]*url.URL, 0, len(identities))
		for _, identity := range identities {
			uri, parseErr := url.Parse(identity)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			uris = append(uris, uri)
		}
		certificate := &x509.Certificate{Subject: pkix.Name{CommonName: "test-workload"}, URIs: uris}
		request.TLS = &tls.ConnectionState{VerifiedChains: [][]*x509.Certificate{{certificate}}}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
