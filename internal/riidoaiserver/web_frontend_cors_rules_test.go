package riidoaiserver

import (
	"net/http"
	"reflect"
	"testing"
)

func TestWebFrontendCORSRuleHelpers(t *testing.T) {
	origins := normalizeWebAllowedOrigins([]string{
		" https://console.riido.io ",
		"",
		"https://console.riido.io",
		"http://localhost:5173",
	})
	if want := []string{"https://console.riido.io", "http://localhost:5173"}; !reflect.DeepEqual(origins, want) {
		t.Fatalf("origins = %#v, want %#v", origins, want)
	}

	for _, method := range []string{" get ", "POST", "patch", "DELETE"} {
		if !webCORSMethodAllowed(method) {
			t.Fatalf("method %q should be allowed", method)
		}
	}
	for _, method := range []string{"OPTIONS", "PUT", ""} {
		if webCORSMethodAllowed(method) {
			t.Fatalf("method %q should be rejected", method)
		}
	}

	for _, headers := range []string{
		"",
		"Authorization, Content-Type",
		"x-riido-ai-agent-token, last-event-id, accept",
	} {
		if !webCORSHeadersAllowed(headers) {
			t.Fatalf("headers %q should be allowed", headers)
		}
	}
	if webCORSHeadersAllowed("Authorization, X-Private-Debug") {
		t.Fatal("private debug header should be rejected")
	}
}

func TestAddVaryHeaderDeduplicatesCaseInsensitiveValues(t *testing.T) {
	header := http.Header{}
	header.Add("Vary", "Origin, Access-Control-Request-Method")

	addVaryHeader(header, "origin")
	addVaryHeader(header, "Access-Control-Request-Headers")

	values := header.Values("Vary")
	if len(values) != 2 {
		t.Fatalf("vary values = %#v", values)
	}
	if values[1] != "Access-Control-Request-Headers" {
		t.Fatalf("new vary value = %q", values[1])
	}
}
