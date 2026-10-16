package riidoaiserver

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"
)

const externalAuthorizerSlowThreshold = time.Second

type externalAuthorizerObservation struct{ inFlight atomic.Int64 }

func (o *externalAuthorizerObservation) start(
	ctx context.Context, req AuthorizationRequest, timeout time.Duration,
) (context.Context, func(error)) {
	started := time.Now()
	inFlight := o.inFlight.Add(1)
	ctx, span := StartTraceSpan(ctx, nil, TraceSpanStart{
		Name: "external_authorizer.authorize", Kind: TraceSpanKindClient,
		Attributes: []TraceAttribute{
			StringTraceAttribute("riido.external_authorizer.resource", string(req.Resource)),
			StringTraceAttribute("riido.external_authorizer.action", string(req.Action)),
			Int64TraceAttribute("riido.external_authorizer.in_flight", inFlight),
		},
	})
	return ctx, func(err error) {
		duration := time.Since(started)
		remaining := o.inFlight.Add(-1)
		failureKind, responseError, upstreamStatus := externalAuthorizerFailureFields(err)
		span.SetAttributes(
			StringTraceAttribute("riido.external_authorizer.failure_kind", failureKind),
			StringTraceAttribute("riido.external_authorizer.response_error", responseError),
			Int64TraceAttribute("riido.external_authorizer.upstream_http_status", int64(upstreamStatus)),
			StringTraceAttribute("riido.external_authorizer.outcome", externalAuthorizerOutcome(err)),
			Int64TraceAttribute("riido.external_authorizer.duration_ms", duration.Milliseconds()),
			Int64TraceAttribute("riido.external_authorizer.in_flight_remaining", remaining),
			Int64TraceAttribute("riido.external_authorizer.timeout_ms", timeout.Milliseconds()),
		)
		FinishTraceSpan(span, err)
		if duration >= externalAuthorizerSlowThreshold || externalAuthorizerLogFailure(err) {
			outcome := externalAuthorizerOutcome(err)
			log.Printf("event=external_authorizer_request outcome=%q error_type=%q failure_kind=%q response_error=%q upstream_http_status=%d failure_stage=%q retryable=%t resource=%q action=%q duration_ms=%d in_flight_at_start=%d in_flight_remaining=%d timeout_ms=%d", outcome, outcome, failureKind, responseError, upstreamStatus, "authorize", externalAuthorizerRetryable(err), req.Resource, req.Action, duration.Milliseconds(), inFlight, remaining, timeout.Milliseconds())
		}
	}
}

func externalAuthorizerRetryable(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) ||
		(err != nil && !errors.Is(err, ErrAuthorizationForbidden) &&
			!errors.Is(err, ErrAuthorizationUnauthenticated) && !errors.Is(err, context.Canceled))
}

func externalAuthorizerOutcome(err error) string {
	switch {
	case err == nil:
		return "allowed"
	case errors.Is(err, ErrAuthorizationForbidden):
		return "forbidden"
	case errors.Is(err, ErrAuthorizationUnauthenticated):
		return "unauthenticated"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	default:
		return "error"
	}
}

func externalAuthorizerLogFailure(err error) bool {
	return err != nil && !errors.Is(err, ErrAuthorizationForbidden) &&
		!errors.Is(err, ErrAuthorizationUnauthenticated) && !errors.Is(err, context.Canceled)
}
