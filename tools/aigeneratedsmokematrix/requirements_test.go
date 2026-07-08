package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/requirements"
)

func TestRequirementsConstantsKeepManifestIdentity(t *testing.T) {
	if requirements.DefaultManifest == "" ||
		!strings.Contains(requirements.ExpectedID, "smoke-matrix") {
		t.Fatalf("requirements constants no longer describe the smoke matrix gate")
	}
}
