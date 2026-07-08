package setutil

import "testing"

func TestStringSetKeepsUniqueItems(t *testing.T) {
	got := StringSet([]string{"a", "b", "a"})
	if len(got) != 2 || !got["a"] || !got["b"] {
		t.Fatalf("set = %+v", got)
	}
}
