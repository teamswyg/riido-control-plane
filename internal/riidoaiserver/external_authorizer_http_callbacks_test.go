package riidoaiserver

import (
	"errors"
	"net/http/httptrace"
	"strings"
	"sync"
	"testing"
)

func TestExternalAuthorizerHTTPCallbacksArePrivateAndConcurrent(t *testing.T) {
	o := newExternalAuthorizerHTTPObservation()
	trace := o.trace()
	trace.GetConn("private-host-canary")
	trace.GotConn(httptrace.GotConnInfo{Reused: true})
	trace.WroteRequest(httptrace.WroteRequestInfo{})
	trace.GotFirstResponseByte()
	o.mark("response_body", "")
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			trace.DNSStart(httptrace.DNSStartInfo{Host: "private-dns-canary"})
			trace.ConnectStart("tcp", "private-address-canary")
			trace.ConnectDone("tcp", "private-address-canary", errors.New("private-error-canary"))
			_, _, _ = o.fields()
		})
	}
	wg.Wait()
	phase, reused, attrs := o.fields()
	if phase != "response_body" || reused != 1 {
		t.Fatal("losing dial regressed selected connection phase")
	}
	for _, attr := range attrs {
		if strings.Contains(attr.StringValue(), "canary") {
			t.Fatal("callback retained private material")
		}
	}
	trace.GetConn("private-next-host-canary")
	trace.ConnectStart("tcp", "private-next-address-canary")
	phase, reused, _ = o.fields()
	if phase != "connect" || reused != -1 {
		t.Fatal("reacquired connection inherited the previous response phase")
	}
	unknown := newExternalAuthorizerHTTPObservation()
	phase, _, attrs = unknown.fields()
	if phase != "unobserved" {
		t.Fatal("custom transport must remain unobserved")
	}
	for _, attr := range attrs[2:] {
		if attr.Int64Value != -1 {
			t.Fatal("missing timing must remain unknown")
		}
	}
}
