package main

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

const (
	reactQueryGenCoreHash  = "19638a0317d0f259f08ee00838b02b2487051e87547ba403b088f4363adedce2"
	reactQueryGenReactHash = "39cade6d9b62a55dc233c93d85a26317a0051c57017ed3acf6f224cb01201a69"
)

func TestReactQueryGenBehaviorGolden(t *testing.T) {
	spec := loadTestOpenAPI(t)
	core, err := generate(spec)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	react, err := generateReact(spec)
	if err != nil {
		t.Fatalf("generateReact: %v", err)
	}
	got := reactQueryGenGolden{
		Operations: len(flattenOperations(spec.Paths)),
		Schemas:    len(spec.Components.Schemas),
		CoreBytes:  len(core),
		ReactBytes: len(react),
		CoreHash:   sha256String(core),
		ReactHash:  sha256String(react),
	}
	want := reactQueryGenGolden{
		Operations: 57,
		Schemas:    65,
		CoreBytes:  375593,
		ReactBytes: 101874,
		CoreHash:   reactQueryGenCoreHash,
		ReactHash:  reactQueryGenReactHash,
	}
	if got != want {
		t.Fatalf("reactquerygen behavior drifted\ngot:  %+v\nwant: %+v", got, want)
	}
}

type reactQueryGenGolden struct {
	Operations int
	Schemas    int
	CoreBytes  int
	ReactBytes int
	CoreHash   string
	ReactHash  string
}

func sha256String(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:])
}
