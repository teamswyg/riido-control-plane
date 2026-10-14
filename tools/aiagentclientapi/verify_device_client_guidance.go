package main

import (
	"fmt"
	"strings"
)

func verifyDeviceClientGuidance(guidance deviceClientGuidance) error {
	if strings.TrimSpace(guidance.ReadEndpoint) == "" || strings.TrimSpace(guidance.Purpose) == "" {
		return fmt.Errorf("device client guidance endpoint and purpose are required")
	}
	if len(guidance.Rules) < 8 {
		return fmt.Errorf("device client guidance requires all state rules")
	}
	for _, marker := range []string{"is_owned_by_viewer", "update_required", "offline", "version_unknown", "ready", "download_url"} {
		found := false
		for _, rule := range guidance.Rules {
			found = found || strings.Contains(rule, marker)
		}
		if !found {
			return fmt.Errorf("device client guidance rule %q is required", marker)
		}
	}
	return nil
}
