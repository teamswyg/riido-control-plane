package riidoaiserver

import (
	"errors"
	"net/url"
	"strings"
)

func normalizeAgentProfileThumbnailURL(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", errors.New("profile_thumbnail_url must be an https URL")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("profile_thumbnail_url must not include query or fragment")
	}
	return trimmed, nil
}
