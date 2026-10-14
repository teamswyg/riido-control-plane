package riidoaiserver

import (
	"net/http"
	"strings"
)

func requestToken(r *http.Request) (string, bool) {
	if token := strings.TrimSpace(r.Header.Get(aiAgentTokenHeader)); token != "" {
		return token, true
	}
	got := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(got, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(got, "Bearer "))
	if token == "" {
		return "", false
	}
	return token, true
}
