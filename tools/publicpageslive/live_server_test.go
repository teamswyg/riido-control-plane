package main

import (
	"net/http"
	"net/http/httptest"
)

type liveCase struct {
	html   string
	status string
	pages  string
	badge  string
	code   int
}

func liveServer(tc liveCase) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tc.code != 0 && r.URL.Path == "/status.json" {
			http.Error(w, "bad status", tc.code)
			return
		}
		switch r.URL.Path {
		case "/":
			_, _ = w.Write([]byte(defaultText(tc.html, "<html>ok</html>")))
		case "/status.json":
			_, _ = w.Write([]byte(defaultText(tc.status, validStatusJSON())))
		case "/pages-status.json":
			_, _ = w.Write([]byte(defaultText(tc.pages, validPagesJSON())))
		case "/status-badge.json":
			_, _ = w.Write([]byte(defaultText(tc.badge, validBadgeJSON())))
		default:
			http.NotFound(w, r)
		}
	}))
}

func defaultText(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
