package riidoaiserver

import (
	"log"
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
		defer func() {
			statusCode := recorder.statusCode
			recovered := recover()
			if recovered != nil {
				statusCode = http.StatusInternalServerError
			}
			route := httpMetricRoute(r.Method, r.URL.Path, r.Pattern, statusCode)
			metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
				Stream:        strings.HasPrefix(recorder.Header().Get("Content-Type"), "text/event-stream"),
				Method:        r.Method,
				Route:         route,
				ClientSurface: traceHTTPClientSurface(route, r.URL.Path, r.UserAgent()),
				StatusCode:    statusCode,
				Duration:      time.Since(startedAt),
			})
			if statusCode >= http.StatusInternalServerError {
				log.Printf(
					"event=http_request_failed method=%q route=%q client_surface=%q status_code=%d duration_ms=%d trace_id=%q span_id=%q",
					r.Method, route, traceHTTPClientSurface(route, r.URL.Path, r.UserAgent()),
					statusCode, time.Since(startedAt).Milliseconds(), requestTraceID(r), requestSpanID(r),
				)
			}
			if recovered != nil {
				panic(recovered)
			}
		}()
		next.ServeHTTP(recorder, r)
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
