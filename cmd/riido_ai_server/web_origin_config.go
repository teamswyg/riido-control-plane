package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func webAllowedOriginsFromEnv() ([]string, error) {
	return parseWebAllowedOrigins(os.Getenv(envWebAllowedOrigins))
}

func parseWebAllowedOrigins(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	seen := map[string]struct{}{}
	var origins []string
	for part := range strings.SplitSeq(raw, ",") {
		origin, err := normalizeWebOrigin(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", envWebAllowedOrigins, err)
		}
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	return origins, nil
}

func normalizeWebOrigin(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if raw == "*" {
		return "", errors.New("wildcard origin is not supported")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse origin %q: %w", raw, err)
	}
	return normalizedParsedWebOrigin(raw, parsed)
}
