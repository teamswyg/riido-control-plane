package riidoaiserver

import "strings"

func clientVisibleProviderLimitMessage(message string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return "", false
	}
	for _, marker := range providerLimitMessageMarkers() {
		if strings.Contains(normalized, marker) {
			return clientMessageCloudCreditInsufficient, true
		}
	}
	return "", false
}

func providerLimitMessageMarkers() []string {
	return []string{
		"token usage limit exceeded",
		"token limit exceeded",
		"token quota exceeded",
		"quota exceeded",
		"insufficient credits",
		"insufficient credit",
		"credit limit exceeded",
		"cloud ai",
		"토큰 이용 한도 초과",
		"토큰 사용 한도 초과",
		"크레딧이 부족",
	}
}
