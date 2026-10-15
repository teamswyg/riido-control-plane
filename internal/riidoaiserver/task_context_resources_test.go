package riidoaiserver

import (
	"strings"
	"testing"
)

func TestTaskContextResourceIdentityAndIncompleteURLsStayExplicit(t *testing.T) {
	data := []byte(`{"files":[{"id":"file-id","fileName":"acceptance.pdf","downloadUrl":"https://cdn.example.test/file?token=fixture","fileLocation":"https://cdn.example.test/public.pdf"}]}`)
	resources, complete := decodePrivateContextResources(data, "file")
	if complete || len(resources) != 1 || resources[0].ID != "file-id" || resources[0].URL != "https://cdn.example.test/public.pdf" {
		t.Fatalf("resource normalization: %+v complete=%v", resources, complete)
	}
	var out strings.Builder
	writeTaskContextResources(&out, resources, "partial", 2500)
	if strings.Contains(out.String(), "token=fixture") || !strings.Contains(out.String(), "file-id") || !strings.Contains(out.String(), "not been downloaded") || !strings.Contains(out.String(), "incomplete") {
		t.Fatalf("unsafe/completeness metadata = %q", out.String())
	}
}

func TestTaskContextFallbackDocumentDoesNotClaimFreshness(t *testing.T) {
	var out strings.Builder
	writeTaskContextDocumentSource(&out, AIAgentTaskContextDocument{Source: "component_snapshot"})
	if !strings.Contains(out.String(), "freshness is unconfirmed") {
		t.Fatal("fallback freshness hidden")
	}
}
