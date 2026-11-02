package riidoaiserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type requestCorrelationKey struct{}

// Generate on the server, including authorization failures, independent of tracing availability.
func withRequestCorrelation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bytes [16]byte
		if _, err := rand.Read(bytes[:]); err == nil {
			id := hex.EncodeToString(bytes[:])
			w.Header().Set("X-Riido-Request-Id", id)
			r = r.WithContext(context.WithValue(r.Context(), requestCorrelationKey{}, id))
		}
		next.ServeHTTP(w, r)
	})
}

func requestCorrelationID(r *http.Request) string {
	id, _ := r.Context().Value(requestCorrelationKey{}).(string)
	return id
}
