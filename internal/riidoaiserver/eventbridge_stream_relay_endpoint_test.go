package riidoaiserver

import (
	"strings"
	"testing"
)

func TestNormalizeEventBridgeEndpointBoundaries(t *testing.T) {
	endpoint, host, err := normalizeEventBridgeEndpoint("ap-northeast-2", "")
	if err != nil {
		t.Fatalf("normalize default: %v", err)
	}
	if endpoint != "https://events.ap-northeast-2.amazonaws.com" || host != "events.ap-northeast-2.amazonaws.com" {
		t.Fatalf("default endpoint=%q host=%q", endpoint, host)
	}
	cases := []struct {
		name     string
		endpoint string
		want     string
	}{
		{"scheme", "ftp://events.ap-northeast-2.amazonaws.com", "http or https"},
		{"host", "https:///events", "host is required"},
		{"query", "https://events.ap-northeast-2.amazonaws.com?x=1", "query"},
		{"fragment", "https://events.ap-northeast-2.amazonaws.com#x", "fragment"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := normalizeEventBridgeEndpoint("ap-northeast-2", tc.endpoint)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("normalize error = %v, want %q", err, tc.want)
			}
		})
	}
}
