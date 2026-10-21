package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// CreateOnboardingAgentRequest adds a retry key only to fixture creation. Normal
// agent creation continues to treat every request as an independent intent.
type CreateOnboardingAgentRequest struct {
	CreateAgentConfigurationRequest
	RequestID json.RawMessage `json:"request_id,omitempty"`
}

type onboardingAgentCreateRequestIDContextKey struct{}

type onboardingAgentCreateMutationContextKey struct{}

const (
	onboardingAgentCreateRetention    = 24 * time.Hour
	onboardingAgentCreateReceiptBytes = 128 * 1024
)

var errOnboardingAgentCreateConflict = errors.New("onboarding agent creation conflict")

type onboardingAgentCreateDependencyError struct{ cause error }

func (e onboardingAgentCreateDependencyError) Error() string {
	return "onboarding agent creation unavailable"
}
func (e onboardingAgentCreateDependencyError) Unwrap() error { return e.cause }

type onboardingAgentCreateReceipt struct {
	Fingerprint    string    `json:"fingerprint"`
	AgentID        string    `json:"agent_id"`
	AgentCreatedAt time.Time `json:"agent_created_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func withOnboardingAgentCreateRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, onboardingAgentCreateRequestIDContextKey{}, requestID)
}
