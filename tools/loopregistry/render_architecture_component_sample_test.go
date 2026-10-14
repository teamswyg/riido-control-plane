package main

import (
	"strings"
	"testing"
)

func TestArchitectureComponentDocSampleBoundsValues(t *testing.T) {
	got := architectureComponentDocSample([]string{"claim-a", "claim-b", "claim-c"})
	want := "`claim-a`<br>`claim-b`<br>+1"
	if got != want {
		t.Fatalf("architectureComponentDocSample() = %q, want %q", got, want)
	}
}

func TestArchitectureComponentDocSampleEscapesMarkdownCells(t *testing.T) {
	got := architectureComponentDocSample([]string{"go test ./x | tee out"})
	want := "`go test ./x \\| tee out`"
	if got != want {
		t.Fatalf("architectureComponentDocSample() = %q, want %q", got, want)
	}
}

func TestArchitectureComponentDocSampleTruncatesLongValues(t *testing.T) {
	got := architectureComponentDocSample([]string{strings.Repeat("0123456789", 10)})
	want := "`" + strings.Repeat("0123456789", 9) + "012345...`"
	if got != want {
		t.Fatalf("architectureComponentDocSample() = %q, want %q", got, want)
	}
}
