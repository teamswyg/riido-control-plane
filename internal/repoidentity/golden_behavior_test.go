package repoidentity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

const repoIdentityGoldenSHA256 = "12c661b639792ddf9d87493bd9383880e5cc43878d68515069d3f8666f36b2ff"

type repoIdentityGolden struct {
	Name       string `json:"name"`
	ModulePath string `json:"module_path"`
	Boundary   string `json:"boundary"`
}

func TestRepoIdentityBehaviorGolden(t *testing.T) {
	body, err := json.Marshal(repoIdentityGolden{
		Name:       Name,
		ModulePath: ModulePath,
		Boundary:   Boundary,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256Hex(body); got != repoIdentityGoldenSHA256 {
		t.Fatalf("repo identity golden hash = %s", got)
	}
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
