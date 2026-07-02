package riidoaiserver

import "strings"

var clientVisibleRewriteLowerMarkers = []string{
	"/users", "/private/var", "/var/folders", "/tmp/",
	"approval", "permission", "allow /", "bash)", "write)",
	"token", "quota", "credit", "cloud ai", "context ",
	"권한", "승인", "터미널에서 직접", "go 명령", "토큰", "크레딧",
}

func clientVisibleTaskThreadTextIsPlain(message string) bool {
	lower := strings.ToLower(message)
	if clientVisibleTaskThreadTextNeedsRewrite(message, lower) {
		return false
	}
	return clientVisibleLocalizedTaskThreadText(message) == message
}

func clientVisibleTaskThreadTextNeedsRewrite(message, lower string) bool {
	if strings.ContainsAny(message, "<>[]()`") {
		return true
	}
	for _, marker := range clientVisibleRewriteLowerMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
