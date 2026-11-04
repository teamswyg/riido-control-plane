package authpep

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NewCLIThreadReadAuthorizer(config CLIThreadReadConfig) (*CLIThreadReadAuthorizer, error) {
	if !cliThreadReadOrigin(config.MonolithOrigin) || !cliThreadReadOrigin(config.Resource) || config.MonolithOrigin == config.Resource || len(config.ServiceToken) < 32 || len(config.ServiceToken) > 512 {
		return nil, errCLIThreadReadUnavailable
	}
	for _, c := range config.ServiceToken {
		if c < 33 || c > 126 {
			return nil, errCLIThreadReadUnavailable
		}
	}
	client := http.Client{}
	if config.HTTPClient != nil {
		client = *config.HTTPClient
	}
	client.Jar = nil
	client.Timeout = 5 * time.Second
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &CLIThreadReadAuthorizer{origin: config.MonolithOrigin, resource: config.Resource, token: config.ServiceToken, client: &client, now: time.Now}, nil
}

func cliThreadReadOrigin(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && len(raw) <= 1024 && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Path == "" && u.RawPath == "" && u.RawQuery == "" && !u.ForceQuery && u.Fragment == "" && raw == "https://"+u.Host
}

func cliThreadReadID(value string) bool {
	if value == "" || len(value) > 512 || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if r < 33 || r == 127 || strings.ContainsRune("/\\?#", r) {
			return false
		}
	}
	return true
}
