package authpep

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestCLIThreadReadClientNeverFollowsRedirectOrCopiesCookies(t *testing.T) {
	var leaked atomic.Int64
	redirect := ""
	_, server := newThreadReadFixture(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/leak" {
			leaked.Add(1)
			return
		}
		if r.Header.Get("Cookie") != "" {
			t.Error("cookie crossed service boundary")
		}
		w.Header().Set("Location", redirect)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	redirect = server.URL + "/leak"
	client := server.Client()
	client.Jar, _ = cookiejar.New(nil)
	u, _ := url.Parse(server.URL)
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "session", Value: "fixture"}})
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return nil }
	config := CLIThreadReadConfig{MonolithOrigin: server.URL, Resource: threadReadFixtureResource, ServiceToken: threadReadFixtureService, HTTPClient: client}
	a, err := NewCLIThreadReadAuthorizer(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Authorize(t.Context(), threadReadFixtureToken(), threadReadFixtureRequest()); err == nil || leaked.Load() != 0 {
		t.Fatal("redirect accepted", err, leaked.Load())
	}
	raw, _ := json.Marshal(config)
	for _, out := range []string{string(raw), fmt.Sprintf("%+v", config), fmt.Sprintf("%#v", a)} {
		if strings.Contains(out, threadReadFixtureService) {
			t.Fatal("credential leaked")
		}
	}
}

func TestCLIThreadReadClientConfigurationAndRequestScope(t *testing.T) {
	for _, origin := range []string{"", "http://example.test", "https://a:b@example.test", "https://example.test/path", "https://example.test?", "https://example.test#x"} {
		if _, err := NewCLIThreadReadAuthorizer(CLIThreadReadConfig{MonolithOrigin: origin, Resource: threadReadFixtureResource, ServiceToken: threadReadFixtureService}); err == nil {
			t.Fatal("invalid origin accepted")
		}
	}
	req := threadReadFixtureRequest()
	for _, resource := range []riidoaiserver.AuthorizationResource{AuthorizationResourceAIAgentClient, "cli_installation"} {
		req.Resource = resource
		if validCLIThreadReadRequest(threadReadFixtureToken(), req) {
			t.Fatal("generic authority accepted")
		}
	}
	req = threadReadFixtureRequest()
	req.Action = AuthorizationActionCreate
	if validCLIThreadReadRequest(threadReadFixtureToken(), req) {
		t.Fatal("write authority accepted")
	}
}
