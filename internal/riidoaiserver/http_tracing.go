package riidoaiserver

import (
	"errors"
	"log"
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
		requestCtx := ExtractTraceContext(r.Context(), recorder, r.Header)
		ctx, span := StartTraceSpan(requestCtx, recorder, TraceSpanStart{
			Name: "HTTP " + r.Method + " " + route,
			Kind: TraceSpanKindServer,
			Attributes: []TraceAttribute{
				StringTraceAttribute(metadatakeys.HTTPRequestMethod.String(), r.Method),
				StringTraceAttribute(metadatakeys.HTTPRoute.String(), route),
				StringTraceAttribute(metadatakeys.RiidoTraceSurface.String(), "control_plane_http"),
				StringTraceAttribute(riidoClientSurfaceTraceKey, traceHTTPClientSurface(route, r.URL.Path, r.UserAgent())),
			},
		})
		recorderResponse := &httpStatusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		defer func() {
			statusCode := recorderResponse.statusCode
			recovered := recover()
			if recovered != nil {
				statusCode = http.StatusInternalServerError
				log.Printf("event=http_handler_panic route=%q", route)
			}
			span.SetAttributes(
				Int64TraceAttribute(metadatakeys.HTTPResponseStatusCode.String(), int64(statusCode)),
				Int64TraceAttribute(metadatakeys.HTTPStatusCode.String(), int64(statusCode)),
			)
			if recovered != nil {
				span.RecordError(errors.New("http handler panic"))
			} else if statusCode >= 500 {
				span.RecordError(httpStatusTraceError(statusCode))
			}
			span.End()
			if recovered != nil {
				panic(recovered)
			}
		}()
		next.ServeHTTP(recorderResponse, r.WithContext(WithTraceRecorder(ctx, recorder)))
	})
}

type httpStatusTraceError int

func (e httpStatusTraceError) Error() string {
	return "http status " + strconv.Itoa(int(e))
}
