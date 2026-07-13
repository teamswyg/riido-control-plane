package riidoaiserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

const (
	riidoHTTPStreamTraceKey          = "riido.http.stream"
	riidoHTTPTimeToFirstByteTraceKey = "riido.http.time_to_first_byte_ms"
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
				StringTraceAttribute(riidoClientSurfaceTraceKey, traceHTTPClientSurface(route, r.URL.Path, r.UserAgent())),
			},
		})
		startedAt := time.Now()
		spanEnded := false
		finishSpan := func(statusCode int, stream bool) {
			if spanEnded {
				return
			}
			attributes := []TraceAttribute{
				Int64TraceAttribute(metadatakeys.HTTPResponseStatusCode.String(), int64(statusCode)),
				Int64TraceAttribute(metadatakeys.HTTPStatusCode.String(), int64(statusCode)),
			}
			if stream {
				attributes = append(attributes,
					StringTraceAttribute(riidoHTTPStreamTraceKey, "sse"),
					Int64TraceAttribute(riidoHTTPTimeToFirstByteTraceKey, durationMilliseconds(time.Since(startedAt))),
				)
			}
			span.SetAttributes(attributes...)
			if statusCode >= 500 {
				span.RecordError(httpStatusTraceError(statusCode))
			}
			span.End()
			spanEnded = true
		}
		recorderResponse := &httpStatusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		recorderResponse.onCommit = func(statusCode int) {
			if isSSEResponse(recorderResponse.Header()) {
				finishSpan(statusCode, true)
			}
		}
		next.ServeHTTP(recorderResponse, r.WithContext(WithTraceRecorder(ctx, recorder)))
		finishSpan(recorderResponse.statusCode, false)
	})
}

type httpStatusTraceError int

func (e httpStatusTraceError) Error() string {
	return "http status " + strconv.Itoa(int(e))
}
