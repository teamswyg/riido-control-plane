package authpep

import (
	"errors"
	"net/http"
	"time"
)

const (
	cliThreadReadPurpose = "task_threads_read"
	cliThreadReadSchema  = "riido-cli-thread-read.v1"
)

var errCLIThreadReadUnavailable = errors.New("CLI thread authorization dependency unavailable")

type CLIThreadReadConfig struct {
	MonolithOrigin string
	Resource       string
	ServiceToken   string
	HTTPClient     *http.Client
}

func (CLIThreadReadConfig) String() string               { return "CLIThreadReadConfig[redacted]" }
func (CLIThreadReadConfig) GoString() string             { return "CLIThreadReadConfig[redacted]" }
func (CLIThreadReadConfig) MarshalJSON() ([]byte, error) { return []byte(`"redacted"`), nil }

type CLIThreadReadAuthorizer struct {
	origin, resource, token string
	client                  *http.Client
	now                     func() time.Time
}

func (*CLIThreadReadAuthorizer) String() string   { return "CLIThreadReadAuthorizer[redacted]" }
func (*CLIThreadReadAuthorizer) GoString() string { return "CLIThreadReadAuthorizer[redacted]" }

type cliThreadReadResponse struct {
	Schema      string    `json:"schema_version"`
	Allowed     bool      `json:"allowed"`
	PrincipalID string    `json:"principal_id"`
	Roles       []string  `json:"roles"`
	Resource    string    `json:"resource"`
	Purpose     string    `json:"purpose"`
	WorkspaceID string    `json:"workspace_id"`
	TaskID      string    `json:"task_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	CheckedAt   time.Time `json:"checked_at"`
}
