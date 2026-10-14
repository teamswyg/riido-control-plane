package riidoaiserver

import (
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

const providerAuthFailureCategory = "provider_authentication"

func clientVisibleProviderAuthCategory(metadata map[string]string) bool {
	if metadata == nil {
		return false
	}
	return strings.TrimSpace(metadata[metadatakeys.AssignmentFailureCategory.String()]) == providerAuthFailureCategory
}

func clientVisibleProviderAuthMessage(message string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return "", false
	}
	for _, marker := range providerAuthMessageMarkers() {
		if strings.Contains(normalized, marker) {
			return clientMessageProviderAuthFailed, true
		}
	}
	return "", false
}

func providerAuthMessageMarkers() []string {
	return []string{
		"failed to authenticate",
		"invalid authentication credentials",
		"api error: 401",
		"401 invalid authentication",
		"authentication credentials",
	}
}
