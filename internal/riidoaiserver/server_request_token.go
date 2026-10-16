package riidoaiserver

import (
	"net/http"
	"strings"
)

func requestToken(r *http.Request) (string, bool) {
	if token := strings.TrimSpace(r.Header.Get(aiAgentTokenHeader)); token != "" {
		return token, true
	}
	return requestBearerToken(r)
}

// requestBearerToken reads only the Authorization Bearer credential. The viewer's
// Riido JWT is needed for monolith reads, so it must not fall back to the agent token header.
func requestBearerToken(r *http.Request) (string, bool) {
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
