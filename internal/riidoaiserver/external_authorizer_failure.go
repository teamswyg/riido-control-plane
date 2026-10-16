package riidoaiserver

import "errors"

// Keep upstream details inside the cause; telemetry uses only bounded fields.
type externalAuthorizerFailure struct {
	kind, responseError string
	upstreamStatus      int
	cause               error
}

func (e externalAuthorizerFailure) Error() string { return "external authorization dependency failed" }
func (e externalAuthorizerFailure) Unwrap() error { return e.cause }

func externalAuthorizerFailed(kind, responseError string, status int, cause error) error {
	return externalAuthorizerFailure{kind: kind, responseError: responseError, upstreamStatus: status, cause: cause}
}

func externalAuthorizerFailureFields(err error) (string, string, int) {
	if err == nil || errors.Is(err, ErrAuthorizationUnauthenticated) || errors.Is(err, ErrAuthorizationForbidden) {
		return "none", "none", 0
	}
	var failure externalAuthorizerFailure
	if !errors.As(err, &failure) {
		return "unclassified", "none", 0
	}
	kind := failure.kind
	switch kind {
	case "transport", "upstream_http", "invalid_response", "internal":
	default:
		return "unclassified", "none", 0
	}
	detail := "none"
	if kind == "invalid_response" {
		switch failure.responseError {
		case "decode", "schema", "contract":
			detail = failure.responseError
		}
	}
	status := failure.upstreamStatus
	if (kind != "upstream_http" && kind != "invalid_response") || status < 100 || status > 599 {
		status = 0
	}
	return kind, detail, status
}
