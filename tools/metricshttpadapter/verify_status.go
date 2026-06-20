package main

import "fmt"

func verifyStatuses(m manifest, result adapterResult) error {
	got := map[string]int{
		"authorized":         result.AuthorizedStatus,
		"missing_scope":      result.MissingScopeStatus,
		"store_unconfigured": result.UnconfiguredStatus,
	}
	for _, required := range m.RequiredStatuses {
		if got[required.Case] != required.Status {
			return fmt.Errorf("status %s = %d, want %d", required.Case, got[required.Case], required.Status)
		}
	}
	return nil
}
