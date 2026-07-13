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
		streamOpenedAt := time.Time{}
		streamRoute := ""
		streamClientSurface := ""
		recorder.onCommit = func(statusCode int) {
			if !isSSEResponse(recorder.Header()) {
				return
			}
			streamOpenedAt = time.Now()
			streamRoute = httpMetricRoute(r.Method, r.URL.Path, r.Pattern, statusCode)
			streamClientSurface = traceHTTPClientSurface(streamRoute, r.URL.Path, r.UserAgent())
			ttfb := streamOpenedAt.Sub(startedAt)
			metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
				Method:        r.Method,
				Route:         streamRoute,
				ClientSurface: streamClientSurface,
				StatusCode:    statusCode,
				Duration:      ttfb,
				ObservedAt:    streamOpenedAt,
			})
			metrics.ObserveSSEStreamOpen(SSEStreamOpenObservation{
				Route:           streamRoute,
				ClientSurface:   streamClientSurface,
				TimeToFirstByte: ttfb,
				ObservedAt:      streamOpenedAt,
			})
		}
		next.ServeHTTP(recorder, r)
		if !streamOpenedAt.IsZero() {
			metrics.ObserveSSEStreamClose(SSEStreamCloseObservation{
				Route:         streamRoute,
				ClientSurface: streamClientSurface,
				Duration:      time.Since(streamOpenedAt),
			})
			return
		}
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

func isSSEResponse(header http.Header) bool {
	contentType := strings.ToLower(strings.TrimSpace(header.Get("Content-Type")))
	return contentType == "text/event-stream" || strings.HasPrefix(contentType, "text/event-stream;")
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
