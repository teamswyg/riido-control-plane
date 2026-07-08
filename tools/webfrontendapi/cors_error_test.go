package main

import "testing"

func TestVerifyCORSResultRejectsDrift(t *testing.T) {
	tc := corsCase{Name: "case", WantStatus: 204, WantAllowOrigin: "https://app.riido.io"}
	for name, got := range map[string]corsEvidence{
		"status":      {Name: "case", HTTPStatus: 200, AllowOrigin: "https://app.riido.io"},
		"origin":      {Name: "case", HTTPStatus: 204, AllowOrigin: "https://other.example"},
		"credentials": {Name: "case", HTTPStatus: 204, AllowOrigin: "https://app.riido.io", Credentials: "true"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := verifyCORSResult(tc, got); err == nil {
				t.Fatalf("expected CORS verification error")
			}
		})
	}
	tc.WantCredentials = "true"
	if err := verifyCORSResult(tc, corsEvidence{Name: "case", HTTPStatus: 204, AllowOrigin: tc.WantAllowOrigin}); err == nil {
		t.Fatalf("expected missing credentials error")
	}
}

func TestVerifyCORSCasesRejectsRuntimeDrift(t *testing.T) {
	m := webFrontendAPIFixture()
	m.CORSCases[0].WantStatus = 500
	if _, err := verifyCORSCases(m.CORSCases); err == nil {
		t.Fatalf("expected CORS case error")
	}
}
