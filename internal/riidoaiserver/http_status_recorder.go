package riidoaiserver

import "net/http"

type httpStatusRecorder struct {
	http.ResponseWriter
	statusCode int
	wrote      bool
	onCommit   func(int)
}

func (r *httpStatusRecorder) WriteHeader(statusCode int) {
	if r.wrote {
		return
	}
	r.wrote = true
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
	if r.onCommit != nil {
		r.onCommit(statusCode)
	}
}

func (r *httpStatusRecorder) Write(p []byte) (int, error) {
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(p)
}

func (r *httpStatusRecorder) Flush() {
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *httpStatusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
