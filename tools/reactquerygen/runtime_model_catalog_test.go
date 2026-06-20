package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIClientDeliveryUsesResolvedRuntimeModelCatalogWording(t *testing.T) {
	docPath := filepath.Join("..", "..", "docs", "30-architecture", "api-client-delivery.md")
	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read api client delivery doc: %v", err)
	}
	body := string(data)
	for _, forbidden := range runtimeModelCatalogForbiddenPhrases {
		if strings.Contains(body, forbidden) {
			t.Fatalf("api client delivery doc still uses unresolved runtime model catalog wording %q", forbidden)
		}
	}
	for _, required := range runtimeModelCatalogRequiredPhrases {
		if !strings.Contains(body, required) {
			t.Fatalf("api client delivery doc missing resolved runtime model catalog wording %q", required)
		}
	}
}
