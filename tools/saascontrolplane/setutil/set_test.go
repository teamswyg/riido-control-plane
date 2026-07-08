package setutil

import "testing"

func TestStringSet(t *testing.T) {
	got := StringSet([]string{"alpha", "beta", "alpha"})
	if len(got) != 2 || !got["alpha"] || !got["beta"] {
		t.Fatalf("set = %#v", got)
	}
}
