package riidoaiserver

import (
	"errors"
	"log"
	"strings"
)

const (
	pollErrorCategoryAgentNotRegistered = "binding_agent_not_registered"
	pollErrorCategoryDaemonMissing      = "binding_daemon_id_missing"
	pollErrorCategoryRuntimeMissing     = "binding_runtime_id_missing"
	pollErrorCategoryDaemonMismatch     = "binding_daemon_id_mismatch"
	pollErrorCategoryDeviceMismatch     = "binding_device_id_mismatch"
	pollErrorCategoryRuntimeMismatch    = "binding_runtime_id_mismatch"
	pollErrorCategoryBindingValidation  = "binding_validation"
	pollErrorCategoryBadRequest         = "bad_request"
)

func logAgentPollRejected(agentID string, err error) {
	log.Printf(
		"event=agent_poll_rejected route=/v1/agents/{agent_id}/poll status=400 agent_id=%q poll_error_category=%q",
		strings.TrimSpace(agentID),
		pollErrorCategory(err),
	)
}

func pollErrorCategory(err error) string {
	if err == nil {
		return pollErrorCategoryBadRequest
	}
	message := err.Error()
	switch {
	case strings.Contains(message, "is not registered"):
		return pollErrorCategoryAgentNotRegistered
	case strings.Contains(message, "daemon_id is required"):
		return pollErrorCategoryDaemonMissing
	case strings.Contains(message, "runtime_id is required"):
		return pollErrorCategoryRuntimeMissing
	case strings.Contains(message, "bound to daemon_id"):
		return pollErrorCategoryDaemonMismatch
	case strings.Contains(message, "bound to device_id"):
		return pollErrorCategoryDeviceMismatch
	case strings.Contains(message, "bound to runtime_id"):
		return pollErrorCategoryRuntimeMismatch
	case errors.Is(err, ErrAgentBindingValidation):
		return pollErrorCategoryBindingValidation
	default:
		return pollErrorCategoryBadRequest
	}
}
