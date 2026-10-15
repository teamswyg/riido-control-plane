package main

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

const (
	reactQueryGenCoreHash  = "fdbbfbd37e9bf3069d8d0d0301cbb190f06e29b9dff5a77282f660b1e1a0c6e2"
	reactQueryGenReactHash = "d44cc9a9fd0ec67479b53c9fb65e79332bf467577b8053de1ab832984c607484"
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
		Schemas:    68,
		CoreBytes:  380071,
		ReactBytes: 102106,
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
