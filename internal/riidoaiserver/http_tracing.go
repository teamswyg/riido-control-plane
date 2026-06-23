package riidoaiserver

import (
	"net/http"
	"strconv"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func withHTTPTracing(next http.Handler, recorder TraceRecorder) http.Handler {
	if recorder == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := traceHTTPRoute(r.Method, r.URL.Path)
		if route == "" {
			next.ServeHTTP(w, r)
			return
		}
		ctx, span := StartTraceSpan(r.Context(), recorder, TraceSpanStart{
			Name: "HTTP " + r.Method + " " + route,
			Kind: TraceSpanKindServer,
			Attributes: []TraceAttribute{
				StringTraceAttribute(metadatakeys.HTTPRequestMethod.String(), r.Method),
				StringTraceAttribute(metadatakeys.HTTPRoute.String(), route),
				StringTraceAttribute(metadatakeys.RiidoTraceSurface.String(), "control_plane_http"),
			},
		})
		recorderResponse := &httpStatusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorderResponse, r.WithContext(WithTraceRecorder(ctx, recorder)))
		span.SetAttributes(
			Int64TraceAttribute(metadatakeys.HTTPResponseStatusCode.String(), int64(recorderResponse.statusCode)),
			Int64TraceAttribute(metadatakeys.HTTPStatusCode.String(), int64(recorderResponse.statusCode)),
		)
		if recorderResponse.statusCode >= 500 {
			span.RecordError(httpStatusTraceError(recorderResponse.statusCode))
		}
		span.End()
	})
}

type httpStatusTraceError int

func (e httpStatusTraceError) Error() string {
	return "http status " + strconv.Itoa(int(e))
}
