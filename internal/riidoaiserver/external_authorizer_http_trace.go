package riidoaiserver

import (
	"crypto/tls"
	"net/http/httptrace"
)

func (o *externalAuthorizerHTTPObservation) trace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn:  func(string) { o.beginConnection() },
		DNSStart: func(httptrace.DNSStartInfo) { o.mark("dns", "") },
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if info.Err == nil {
				o.mark("connection_wait", "dns_done_elapsed_ms")
			}
		},
		ConnectStart: func(string, string) { o.mark("connect", "") },
		ConnectDone: func(_, _ string, err error) {
			if err == nil {
				o.mark("connection_wait", "connect_done_elapsed_ms")
			}
		},
		TLSHandshakeStart: func() { o.mark("tls", "") },
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			if err == nil {
				o.mark("connection_wait", "tls_done_elapsed_ms")
			}
		},
		GotConn: func(info httptrace.GotConnInfo) {
			o.mu.Lock()
			o.reused = info.Reused
			o.mu.Unlock()
			o.mark("request_write", "connection_elapsed_ms")
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			if info.Err == nil {
				o.mark("response_headers", "write_elapsed_ms")
			}
		},
		GotFirstResponseByte: func() { o.mark("response_headers", "first_byte_elapsed_ms") },
	}
}
