package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPublicPagesLiveBehaviorGolden(t *testing.T) {
	server := testServer(t, false)
	var out bytes.Buffer
	err := run(
		[]string{"-base-url", server.URL, "-out", "-"},
		&out,
		time.Date(2026, 7, 1, 1, 2, 3, 0, time.UTC),
		server.Client(),
	)
	if err != nil {
		t.Fatal(err)
	}
	body := strings.ReplaceAll(out.String(), server.URL, "https://fixture.invalid")
	got := fmt.Sprintf("%x", sha256.Sum256([]byte(body)))
	const want = "b232fc4f78a64f2434b17cae957ae0be585244b064d6cf73dfa48e3eb69414a6"
	if got != want {
		t.Fatalf("public pages live golden sha = %s\n%s", got, body)
	}
}
