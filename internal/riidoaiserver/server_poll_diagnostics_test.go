package riidoaiserver

import (
	"errors"
	"testing"
)

func TestPollErrorCategoryClassifiesBindingFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "unregistered", err: newAgentBindingValidationErrorf("agent agent-a is not registered"), want: pollErrorCategoryAgentNotRegistered},
		{name: "daemon missing", err: newAgentBindingValidationErrorf("daemon_id is required"), want: pollErrorCategoryDaemonMissing},
		{name: "runtime missing", err: newAgentBindingValidationErrorf("runtime_id is required"), want: pollErrorCategoryRuntimeMissing},
		{name: "daemon mismatch", err: newAgentBindingValidationErrorf("agent agent-a is bound to daemon_id daemon-a"), want: pollErrorCategoryDaemonMismatch},
		{name: "device mismatch", err: newAgentBindingValidationErrorf("agent agent-a is bound to device_id device-a"), want: pollErrorCategoryDeviceMismatch},
		{name: "runtime mismatch", err: newAgentBindingValidationErrorf("agent agent-a is bound to runtime_id runtime-a"), want: pollErrorCategoryRuntimeMismatch},
		{name: "generic binding", err: agentBindingValidationError{message: "binding rejected"}, want: pollErrorCategoryBindingValidation},
		{name: "generic bad request", err: errors.New("invalid json"), want: pollErrorCategoryBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pollErrorCategory(tt.err); got != tt.want {
				t.Fatalf("pollErrorCategory() = %q, want %q", got, tt.want)
			}
		})
	}
}
