package controlplane

import (
	"io"
	"net/http"
	"strings"
)

const controlPlaneOwnerRequestLimit = 16 << 10

type handler struct{ useCase UseCase }

// NewGraphQLHandler returns the source-ready owner receiver. Production does not register it yet.
func NewGraphQLHandler(useCase UseCase) http.Handler {
	return handler{useCase: useCase}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeControlPlaneOwnerFailure(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.useCase == nil {
		writeControlPlaneOwnerFailure(w, http.StatusServiceUnavailable, "control-plane owner receiver is unavailable")
		return
	}
	mediaType := strings.ToLower(strings.TrimSpace(strings.SplitN(r.Header.Get("Content-Type"), ";", 2)[0]))
	if mediaType != "application/json" {
		writeControlPlaneOwnerFailure(w, http.StatusUnsupportedMediaType, "content type must be application/json")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, controlPlaneOwnerRequestLimit+1))
	operation, decodeErr := decodeControlPlaneOwnerRequest(raw)
	if err != nil || decodeErr != nil || len(raw) > controlPlaneOwnerRequestLimit {
		writeControlPlaneOwnerFailure(w, http.StatusUnprocessableEntity, "control-plane owner request is not registered")
		return
	}
	switch operation.Coordinate {
	case "Query.healthCheck":
		status, callErr := h.useCase.HealthCheck(r.Context())
		if callErr != nil || status != http.StatusOK {
			writeControlPlaneOwnerFailure(w, http.StatusOK, "control-plane healthCheck failed closed")
			return
		}
		writeResponse(w, healthResponse{Data: healthData{HealthCheck: status}})
	case "Query.fireError":
		_ = h.useCase.FireError(r.Context())
		writeResponse(w, fireResponse{
			Data: fireData{FireError: nil}, Errors: []graphQLError{{Message: errControlPlaneOwnerFire.Error()}},
		})
	}
}
