package riidoaiserver

import (
	"net/http"
	"strings"
)

const (
	webCORSAllowedMethods = "GET, POST, PATCH, DELETE, OPTIONS"
	webCORSAllowedHeaders = "X-Riido-AI-Agent-Token, Authorization, Content-Type, Accept, Last-Event-ID"
	webCORSMaxAge         = "600"
)

var webCORSAllowedRequestHeaders = map[string]struct{}{
	"accept":                 {},
	"authorization":          {},
	"content-type":           {},
	"last-event-id":          {},
	"x-riido-ai-agent-token": {},
}

func (s Server) withWebFrontendCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}
		if cliInstallationHTTPRoute(http.MethodPost, r.URL.Path) != "" {
			writeError(w, http.StatusForbidden, "CLI installation is not a browser surface")
			return
		}
		if !s.webOriginAllowed(origin) {
			if isCORSPreflight(r) {
				writeError(w, http.StatusForbidden, "origin is not allowed")
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		writeCORSHeaders(w, origin)
		if !isCORSPreflight(r) {
			next.ServeHTTP(w, r)
			return
		}
		if !webCORSMethodAllowed(r.Header.Get("Access-Control-Request-Method")) {
			writeMethodNotAllowed(w)
			return
		}
		if !webCORSHeadersAllowed(r.Header.Get("Access-Control-Request-Headers")) {
			writeError(w, http.StatusBadRequest, "unsupported cors request header")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func normalizeWebAllowedOrigins(origins []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		out = append(out, origin)
	}
	return out
}
