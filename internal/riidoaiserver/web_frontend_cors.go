package riidoaiserver

import (
	"net/http"
	"strings"
)

const (
	webCORSAllowedMethods = "GET, POST, PATCH, DELETE, OPTIONS"
	webCORSAllowedHeaders = "Authorization, Content-Type, Accept, Last-Event-ID"
	webCORSMaxAge         = "600"
)

var webCORSAllowedRequestHeaders = map[string]struct{}{
	"accept":        {},
	"authorization": {},
	"content-type":  {},
	"last-event-id": {},
}

func (s Server) withWebFrontendCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			next.ServeHTTP(w, r)
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

func (s Server) webOriginAllowed(origin string) bool {
	for _, allowed := range s.config.WebAllowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func isCORSPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")) != ""
}

func writeCORSHeaders(w http.ResponseWriter, origin string) {
	header := w.Header()
	header.Set("Access-Control-Allow-Origin", origin)
	header.Set("Access-Control-Allow-Methods", webCORSAllowedMethods)
	header.Set("Access-Control-Allow-Headers", webCORSAllowedHeaders)
	header.Set("Access-Control-Max-Age", webCORSMaxAge)
	addVaryHeader(header, "Origin")
	addVaryHeader(header, "Access-Control-Request-Method")
	addVaryHeader(header, "Access-Control-Request-Headers")
}

func webCORSMethodAllowed(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func webCORSHeadersAllowed(raw string) bool {
	for _, header := range strings.Split(raw, ",") {
		header = strings.ToLower(strings.TrimSpace(header))
		if header == "" {
			continue
		}
		if _, ok := webCORSAllowedRequestHeaders[header]; !ok {
			return false
		}
	}
	return true
}

func addVaryHeader(header http.Header, value string) {
	existing := header.Values("Vary")
	for _, field := range existing {
		for _, part := range strings.Split(field, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}
