package riidoaiserver

import (
	"net/url"
	"strconv"
)

// Accept only the exact HTTPS origin returned by the monolith resource contract.
func cliInstallationResource(value string) bool {
	u, err := url.Parse(value)
	if err != nil || len(value) > 1024 || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Opaque != "" || u.Path != "" || u.RawPath != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.String() != value {
		return false
	}
	if raw := u.Port(); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port < 1 || port > 65535 {
			return false
		}
	}
	return true
}
