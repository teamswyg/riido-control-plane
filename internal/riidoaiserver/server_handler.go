package riidoaiserver

import "net/http"

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /internal/agent-comment-deliveries", s.handleCommentDelivery)
	registerServerRoutes(mux, s)
	var handler http.Handler = mux
	if len(s.config.WebAllowedOrigins) > 0 {
		handler = s.withWebFrontendCORS(handler)
	}
	handler = withHTTPTransactionMetrics(handler, s.config.HTTPTransactions)
	handler = withHTTPTracing(handler, s.config.TraceRecorder)
	return handler
}
