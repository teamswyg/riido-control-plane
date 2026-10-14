package rendertext

import "testing"

func TestCodeList(t *testing.T) {
	got := CodeList([]string{"endpoint", "hook"})
	want := "`endpoint`, `hook`"
	if got != want {
		t.Fatalf("code list = %q, want %q", got, want)
	}
}

func TestEmptyAwareCodeList(t *testing.T) {
	if got := EmptyAwareCodeList(nil); got != "`none`" {
		t.Fatalf("empty list = %q", got)
	}
	if got := EmptyAwareCodeList([]string{"one"}); got != "`one`" {
		t.Fatalf("non-empty list = %q", got)
	}
}
