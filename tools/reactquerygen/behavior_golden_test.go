package main

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

const (
	reactQueryGenCoreHash  = "e2dd7bda135865c5fe77fe06c39d7921455b04e0739eda7b6a0009fa9b0b5b76"
	reactQueryGenReactHash = "209bb6047f33beab6c4770385cbf468726d71df60c80cf86c0013e3003f01b2c"
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
		Operations: 64,
		Schemas:    90,
		CoreBytes:  444518,
		ReactBytes: 114668,
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
