package main

import (
	"fmt"
	"net/url"
)

func normalizedParsedWebOrigin(raw string, parsed *url.URL) (string, error) {
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("origin %q must use http or https", raw)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("origin %q must include a host", raw)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", fmt.Errorf("origin %q must not include path, query, fragment, or userinfo", raw)
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}
