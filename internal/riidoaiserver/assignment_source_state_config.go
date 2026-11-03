package riidoaiserver

import (
	"errors"
	"net/http"
	"net/url"
)

const AgentSourceStateKeyHeader = "X-Riido-Agent-Source-State-Key"

// assignmentSourceStateLimit matches the monolith lookup limit. A larger set
// is not asked about and stays an unknown omission (full refresh).
const assignmentSourceStateLimit = 32

const (
	AssignmentSourceDeleted    = "deleted"
	AssignmentSourceUnselected = "unselected"
	// AssignmentSourceHidden is stored only for the comparison: a source the
	// requester can no longer read is left out of the result without a reason.
	AssignmentSourceHidden = "hidden"
)

// AgentSourceStateConfig points at the monolith lookup that tells whether a
// source the previous run saw is still readable, deleted or hidden.
type AgentSourceStateConfig struct {
	BaseURL    string
	Key        string
	HTTPClient *http.Client
}

func (config AgentSourceStateConfig) Validate() error {
	if config.BaseURL == "" && config.Key == "" {
		return nil
	}
	endpoint, err := url.Parse(config.BaseURL)
	if err != nil {
		return errors.New("invalid agent source state URL")
	}
	invalidScheme := endpoint.Scheme != "https" && (endpoint.Scheme != "http" || (endpoint.Hostname() != "127.0.0.1" && endpoint.Hostname() != "localhost"))
	if endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" || invalidScheme {
		return errors.New("agent source state URL must use HTTPS (or loopback HTTP)")
	}
	if len(config.Key) < 32 {
		return errors.New("agent source state key must contain at least 32 bytes")
	}
	return nil
}

func (config AgentSourceStateConfig) enabled() bool {
	return config.BaseURL != "" && config.Key != "" && config.Validate() == nil
}
