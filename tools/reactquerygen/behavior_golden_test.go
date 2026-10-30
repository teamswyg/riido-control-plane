package main

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

const (
	reactQueryGenCoreHash  = "b26aafb19c04faad434ca1977dd64bf70181a0610379f075fc907289487a2a49"
	reactQueryGenReactHash = "de2e01620db9b3abb977a42159c3469dd5f7e2864df357cd4ab0decd5d013fe5"
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
		Operations: 66,
		Schemas:    94,
		CoreBytes:  459662,
		ReactBytes: 119092,
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
