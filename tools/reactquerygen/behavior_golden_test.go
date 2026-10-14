package main

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

const (
	reactQueryGenCoreHash  = "15bb8f98e9f84f9069cba8022f43ae2b7052152f641c3f605e2384cfba3a030b"
	reactQueryGenReactHash = "a977089d4f8e9afa66ffb3ce00cf9cd34954471c532725a3940a0089cb8d66a6"
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
		Schemas:    66,
		CoreBytes:  378138,
		ReactBytes: 102060,
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
