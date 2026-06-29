package riidoaiserver

import (
	"net/http"
	"strings"
	"time"
)

func withHTTPTransactionMetrics(next http.Handler, metrics *HTTPTransactionMetrics) http.Handler {
	if metrics == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if excludedHTTPTransactionMetricsPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		startedAt := time.Now()
		recorder := &httpStatusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)
		route := httpMetricRoute(r.Method, r.URL.Path, r.Pattern, recorder.statusCode)
		metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
			Method:        r.Method,
			Route:         route,
			ClientSurface: traceHTTPClientSurface(route, r.URL.Path, r.UserAgent()),
			StatusCode:    recorder.statusCode,
			Duration:      time.Since(startedAt),
		})
	})
}

func httpMetricRoute(method, path, pattern string, statusCode int) string {
	if route := traceHTTPRoute(method, path); route != "" {
		return route
	}
	if pattern = strings.TrimSpace(pattern); pattern != "" {
		return pattern
	}
	if statusCode == http.StatusNotFound {
		return unmatchedHTTPRoute(path)
	}
	return ""
}
