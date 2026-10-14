package layers

import "testing"

func TestIsAllowed(t *testing.T) {
	for _, layer := range []string{"cloud", "contract", "observability"} {
		if !IsAllowed(layer) {
			t.Fatalf("%q should be allowed", layer)
		}
	}
	if IsAllowed("frontend") {
		t.Fatal("unexpected frontend layer")
	}
}

func TestFormatAllowed(t *testing.T) {
	if got := FormatAllowed(); got != "cloud, contract, observability" {
		t.Fatalf("allowed = %q", got)
	}
}
