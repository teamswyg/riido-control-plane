package riidoaiserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const controlPlaneOwnerRequestLimit = 16 << 10

type controlPlaneOwnerHandler struct{ useCase controlPlaneOwnerUseCase }

func newControlPlaneOwnerGraphQLHandler(useCase controlPlaneOwnerUseCase) http.Handler {
	return controlPlaneOwnerHandler{useCase: useCase}
}

func (h controlPlaneOwnerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		writeControlPlaneOwnerResponse(w, map[string]any{"data": map[string]any{"healthCheck": status}})
	case "Query.fireError":
		_ = h.useCase.FireError(r.Context())
		writeControlPlaneOwnerResponse(w, map[string]any{
			"data": map[string]any{"fireError": nil}, "errors": []map[string]string{{"message": errControlPlaneOwnerFire.Error()}},
		})
	}
}

func writeControlPlaneOwnerFailure(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/graphql-response+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"errors": []map[string]string{{"message": message}}})
}

func writeControlPlaneOwnerResponse(w http.ResponseWriter, payload map[string]any) {
	w.Header().Set("Content-Type", "application/graphql-response+json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}
