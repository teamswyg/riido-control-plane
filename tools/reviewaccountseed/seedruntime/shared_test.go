package seedruntime

import "testing"

func TestReviewToken(t *testing.T) {
	if got := ReviewToken(); got != "review-token" {
		t.Fatalf("token = %q", got)
	}
}

func TestSameStrings(t *testing.T) {
	if !SameStrings([]string{"a", "b"}, []string{"a", "b"}) {
		t.Fatal("matching slices should compare equal")
	}
	if SameStrings([]string{"a"}, []string{"a", "b"}) {
		t.Fatal("different lengths should not compare equal")
	}
	if SameStrings([]string{"a", "b"}, []string{"b", "a"}) {
		t.Fatal("different order should not compare equal")
	}
}

func TestTokenHash(t *testing.T) {
	got := tokenHash("review-token")
	want := "3c9269e2a436bba87ad1255617f2231deeb2cb6a63f200e6cffb9c32585f8422"
	if got != want {
		t.Fatalf("hash = %q, want %q", got, want)
	}
}
