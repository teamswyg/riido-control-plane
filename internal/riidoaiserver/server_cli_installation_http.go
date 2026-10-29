package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

func installationHTTP(w http.ResponseWriter, r *http.Request, run func(*http.Request) (any, error)) {
	if len(r.Header.Values("Origin")) > 0 {
		writeCLIInstallationError(w, ErrAuthorizationForbidden)
		return
	}
	defer func() {
		if recover() != nil {
			writeCLIInstallationError(w, errCLIInstallationUnavailable)
		}
	}()
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	result, err := run(r.WithContext(ctx))
	if err != nil {
		writeCLIInstallationError(w, err)
		return
	}
	if err = ctx.Err(); err != nil {
		writeCLIInstallationError(w, err)
		return
	}
	data, err := json.Marshal(result)
	if err != nil || len(data) > 4<<10 {
		writeCLIInstallationError(w, errCLIInstallationUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(append(data, '\n'))
}

func writeCLIInstallationError(w http.ResponseWriter, err error) {
	status, message := http.StatusServiceUnavailable, "CLI installation unavailable"
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		status, message = http.StatusRequestEntityTooLarge, "CLI installation body is too large"
	case errors.Is(err, errCLIInstallationInvalid):
		status, message = http.StatusBadRequest, "invalid CLI installation request"
	case errors.Is(err, ErrAuthorizationUnauthenticated), errors.Is(err, ErrAuthorizationInvalidCredential):
		status, message = http.StatusUnauthorized, "CLI installation authentication required"
	case errors.Is(err, ErrAuthorizationForbidden):
		status, message = http.StatusForbidden, "CLI installation proof is not authorized"
	case errors.Is(err, errCLIInstallationConflict):
		status, message = http.StatusConflict, "CLI installation binding conflicts"
	case errors.Is(err, errCLIInstallationExpired):
		status, message = http.StatusGone, "CLI installation intent expired"
	case errors.Is(err, context.DeadlineExceeded):
		status, message = http.StatusGatewayTimeout, "CLI installation timed out"
	case errors.Is(err, context.Canceled):
		status, message = http.StatusRequestTimeout, "CLI installation cancelled"
	}
	writeError(w, status, message)
}

func (s Server) cliInstallationRepository() (cliInstallationRepository, error) {
	store, ok := s.assignment.(cliInstallationRepository)
	if !ok {
		return nil, errCLIInstallationUnavailable
	}
	return store, nil
}
