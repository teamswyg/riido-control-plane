package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/requirements"
	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/runconfig"
)

func TestAPIClientDeliveryBehaviorGolden(t *testing.T) {
	repo := filepath.Join("..", "..")
	out := filepath.Join(t.TempDir(), "evidence.json")
	err := run(runconfig.Options{
		Repo:        repo,
		Manifest:    requirements.DefaultManifest,
		EvidenceOut: out,
		CheckDoc:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	assertAPIClientDeliveryGolden(t, got)
}
