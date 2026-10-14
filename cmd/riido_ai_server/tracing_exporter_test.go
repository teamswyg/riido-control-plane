package main

import "testing"

func TestOtelTraceHTTPExporterOptionsShape(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantLen  int
	}{
		{name: "plain host", endpoint: "collector", wantLen: 2},
		{name: "http URL", endpoint: "http://collector:4318", wantLen: 2},
		{name: "https URL", endpoint: "https://collector:4318", wantLen: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(otelTraceHTTPExporterOptions(tt.endpoint)); got != tt.wantLen {
				t.Fatalf("options len = %d, want %d", got, tt.wantLen)
			}
		})
	}
}
