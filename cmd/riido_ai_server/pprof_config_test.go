package main

import (
	"strings"
	"testing"
)

func TestConfigFromEnvRestrictsPprofAddrToLoopback(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		want string
	}{
		{name: "localhost", raw: "localhost:6060", want: "localhost:6060"},
		{name: "loopback shorthand", raw: ":6060", want: "127.0.0.1:6060"},
		{name: "ipv6 loopback", raw: "[::1]:6060", want: "[::1]:6060"},
		{name: "disabled", raw: "off", want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clearRiidoAIServerEnv(t)
			t.Setenv(envPprofAddr, tc.raw)
			config, err := configFromEnv()
			if err != nil {
				t.Fatalf("configFromEnv: %v", err)
			}
			if config.PprofAddr != tc.want {
				t.Fatalf("pprof addr = %q, want %q", config.PprofAddr, tc.want)
			}
		})
	}
	for _, raw := range []string{"0.0.0.0:6060", "[::]:6060", "192.0.2.10:6060", "example.com:6060"} {
		t.Run("reject "+raw, func(t *testing.T) {
			clearRiidoAIServerEnv(t)
			t.Setenv(envPprofAddr, raw)
			if _, err := configFromEnv(); err == nil || !strings.Contains(err.Error(), envPprofAddr) {
				t.Fatalf("expected %s rejection for %q, got %v", envPprofAddr, raw, err)
			}
		})
	}
}
