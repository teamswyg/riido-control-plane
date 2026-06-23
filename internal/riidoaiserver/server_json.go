package riidoaiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
)

func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("decode json: trailing data")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, struct {
		SchemaVersion string `json:"schema_version"`
		Error         string `json:"error"`
	}{SchemaVersion: SchemaVersion, Error: message})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="riido_ai_server"`)
	writeError(w, http.StatusUnauthorized, "unauthorized")
}

func copyStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in)+1)
	maps.Copy(out, in)
	return out
}

func writeAIAgentClientError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrAIAgentNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrAIAgentAssigned):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrAIAgentTaskThreadConflict):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusBadRequest, err.Error())
	}
}
