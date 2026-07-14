package riidoaiserver

import "testing"

func TestCompareDaemonVersionsAcceptsReportedLabels(t *testing.T) {
	tests := []struct {
		current  string
		required string
		want     int
		ok       bool
	}{
		{"riido-agentd v0.0.68", "v0.0.68", 0, true},
		{"riido-daemon v0.0.67", "0.0.68", -1, true},
		{"v0.1.0", "v0.0.68", 1, true},
		{"dev", "v0.0.68", 0, false},
	}
	for _, tt := range tests {
		got, ok := compareDaemonVersions(tt.current, tt.required)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("compareDaemonVersions(%q, %q) = %d, %v", tt.current, tt.required, got, ok)
		}
	}
}
