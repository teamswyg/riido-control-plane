package repoidentity

import "testing"

func TestIdentity(t *testing.T) {
	if Name != "riido-control-plane" {
		t.Fatalf("Name = %q", Name)
	}
	if ModulePath != "github.com/teamswyg/riido-control-plane" {
		t.Fatalf("ModulePath = %q", ModulePath)
	}
	if Boundary == "" {
		t.Fatal("Boundary is empty")
	}
}
