package main

import (
	"fmt"
	"net/http/httptest"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func verifyCORSCases(cases []corsCase) ([]corsEvidence, error) {
	results := make([]corsEvidence, 0, len(cases))
	for _, tc := range cases {
		result, err := runCORSCase(tc)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func runCORSCase(tc corsCase) (corsEvidence, error) {
	handler := riidoaiserver.NewServer(riidoaiserver.ServerConfig{
		WebAllowedOrigins: tc.AllowedOrigins,
	}).Handler()
	req := httptest.NewRequest(tc.Method, tc.Path, nil)
	req.Header.Set("Origin", tc.Origin)
	if tc.RequestMethod != "" {
		req.Header.Set("Access-Control-Request-Method", tc.RequestMethod)
	}
	if tc.RequestHeaders != "" {
		req.Header.Set("Access-Control-Request-Headers", tc.RequestHeaders)
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	result := corsResult(tc, resp)
	return result, verifyCORSResult(tc, result)
}

func corsResult(tc corsCase, resp *httptest.ResponseRecorder) corsEvidence {
	return corsEvidence{
		Name:        tc.Name,
		Route:       tc.Method + " " + tc.Path,
		Origin:      tc.Origin,
		HTTPStatus:  resp.Code,
		AllowOrigin: resp.Header().Get("Access-Control-Allow-Origin"),
		Credentials: resp.Header().Get("Access-Control-Allow-Credentials"),
	}
}

func verifyCORSResult(tc corsCase, got corsEvidence) error {
	if got.HTTPStatus != tc.WantStatus {
		return fmt.Errorf("%s status=%d want %d", tc.Name, got.HTTPStatus, tc.WantStatus)
	}
	if got.AllowOrigin != tc.WantAllowOrigin {
		return fmt.Errorf("%s allow-origin=%q want %q", tc.Name, got.AllowOrigin, tc.WantAllowOrigin)
	}
	if tc.WantCredentials != "" && got.Credentials != tc.WantCredentials {
		return fmt.Errorf("%s credentials=%q want %q", tc.Name, got.Credentials, tc.WantCredentials)
	}
	if tc.WantCredentials == "" && got.Credentials != "" {
		return fmt.Errorf("%s unexpected credentials=%q", tc.Name, got.Credentials)
	}
	return nil
}
