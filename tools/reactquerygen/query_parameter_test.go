package main

import (
	"strings"
	"testing"
)

func TestGeneratedDashboardPageReachesURLAndQueryKey(t *testing.T) {
	spec, err := loadOpenAPI("../../contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	coreBytes, err := generate(spec)
	if err != nil {
		t.Fatal(err)
	}
	core := string(coreBytes)
	for _, name := range []string{"getAgentDashboard", "getAgentDashboardV2"} {
		if !strings.Contains(core, "page?: number;") {
			t.Fatal("page type missing")
		}
		start := strings.Index(core, "export async function "+name+"(")
		if start < 0 {
			t.Fatal("operation missing")
		}
		body := core[start:]
		end := strings.Index(body, "\n}\n")
		if end < 0 {
			t.Fatal("generated function has no closing brace")
		}
		body = body[:end+1]
		for _, needle := range []string{`query.set("page", String(params.page))`, "requestPath", "params.page !== undefined"} {
			if !strings.Contains(body, needle) {
				t.Fatalf("%s drops page: %s", name, needle)
			}
		}
		start = strings.Index(core, "export function "+name+"QueryKey(")
		if start < 0 {
			t.Fatal("page key missing")
		}
		body = core[start:]
		end = strings.Index(body, "\n}\n")
		if end < 0 {
			t.Fatal("generated function has no closing brace")
		}
		body = body[:end+1]
		if !strings.Contains(body, "params") {
			t.Fatal("different pages share query key")
		}
	}
}
