package deploypolicy

import (
	"slices"
	"strings"
	"testing"
)

func requireContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("missing %q", want)
	}
}

func requireNotContains(t *testing.T, body, want string) {
	t.Helper()
	if strings.Contains(body, want) {
		t.Fatalf("unexpected %q", want)
	}
}

func requireSliceContains(t *testing.T, items []string, want string) {
	t.Helper()
	if slices.Contains(items, want) {
		return
	}
	t.Fatalf("missing %q in %#v", want, items)
}

func requireStringSetExact(t *testing.T, got, want []string) {
	t.Helper()
	gotSet := map[string]int{}
	for _, item := range got {
		gotSet[item]++
		if gotSet[item] > 1 {
			t.Fatalf("duplicate item %q in %#v", item, got)
		}
	}
	wantSet := map[string]bool{}
	for _, item := range want {
		wantSet[item] = true
		if gotSet[item] == 0 {
			t.Fatalf("missing %q in %#v", item, got)
		}
	}
	for item := range gotSet {
		if !wantSet[item] {
			t.Fatalf("unexpected %q in %#v, expected %#v", item, got, want)
		}
	}
}

func stringSliceContains(items []string, want string) bool {
	return slices.Contains(items, want)
}

func contains(body, want string) bool {
	return strings.Contains(body, want)
}
