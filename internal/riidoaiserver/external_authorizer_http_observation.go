package riidoaiserver

import (
	"sync"
	"time"
)

// HTTP callbacks can overlap (including parallel dials). Retain only finite
// phases, booleans and elapsed times; never retain addresses, headers or errors.
type externalAuthorizerHTTPObservation struct {
	mu      sync.Mutex
	started time.Time
	phase   string
	reused  bool
	timings map[string]int64
}

func newExternalAuthorizerHTTPObservation() *externalAuthorizerHTTPObservation {
	return &externalAuthorizerHTTPObservation{started: time.Now(), phase: "unobserved", timings: map[string]int64{}}
}

func (o *externalAuthorizerHTTPObservation) beginConnection() {
	o.mu.Lock()
	defer o.mu.Unlock()
	// The standard transport may reacquire a connection after a stale idle
	// connection. Do not attribute this attempt to the prior response phase.
	o.phase, o.reused, o.timings = "connection_wait", false, map[string]int64{}
}

func (o *externalAuthorizerHTTPObservation) mark(phase, timing string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if phase != "" {
		// A losing parallel dial may finish after the selected connection has
		// already written/read the response. Do not regress its request phase.
		ranks := map[string]int{"unobserved": 0, "connection_wait": 1, "dns": 2, "connect": 3, "tls": 4, "request_write": 5, "response_headers": 6, "response_body": 7, "complete": 8}
		if ranks[phase] >= ranks[o.phase] {
			o.phase = phase
		}
	}
	if timing != "" {
		o.timings[timing] = time.Since(o.started).Milliseconds()
	}
}

func (o *externalAuthorizerHTTPObservation) fields() (string, int64, []TraceAttribute) {
	o.mu.Lock()
	defer o.mu.Unlock()
	attrs := []TraceAttribute{StringTraceAttribute("riido.external_authorizer.http_phase", o.phase)}
	reused := int64(-1)
	if _, observed := o.timings["connection_elapsed_ms"]; observed {
		reused = 0
		if o.reused {
			reused = 1
		}
	}
	attrs = append(attrs, Int64TraceAttribute("riido.external_authorizer.connection_reused", reused))
	for _, name := range []string{"dns_done_elapsed_ms", "connect_done_elapsed_ms", "tls_done_elapsed_ms", "connection_elapsed_ms", "write_elapsed_ms", "first_byte_elapsed_ms"} {
		value, ok := o.timings[name]
		if !ok {
			value = -1
		}
		attrs = append(attrs, Int64TraceAttribute("riido.external_authorizer."+name, value))
	}
	return o.phase, reused, attrs
}
