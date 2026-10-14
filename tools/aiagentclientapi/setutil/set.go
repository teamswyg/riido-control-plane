package setutil

import "fmt"

func RequireStrings(kind string, got, required []string) error {
	seen := map[string]struct{}{}
	for _, value := range got {
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, ok := seen[value]; !ok {
			return fmt.Errorf("missing %s %q", kind, value)
		}
	}
	return nil
}
