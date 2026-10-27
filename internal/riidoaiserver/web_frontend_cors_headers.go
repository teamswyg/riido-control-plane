package riidoaiserver

import (
	"net/http"
	"slices"
	"strings"
)

func (s Server) webOriginAllowed(origin string) bool {
	return slices.Contains(s.config.WebAllowedOrigins, origin)
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
	for header := range strings.SplitSeq(raw, ",") {
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
		for part := range strings.SplitSeq(field, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}
