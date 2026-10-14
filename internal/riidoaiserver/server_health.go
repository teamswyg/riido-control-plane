package riidoaiserver

import "net/http"

func (s Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, Health{SchemaVersion: SchemaVersion, Status: "ok"})
}

func (s Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, Health{SchemaVersion: SchemaVersion, Status: "ready"})
}
