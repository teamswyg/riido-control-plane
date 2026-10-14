package riidoaiserver

import (
	"errors"
	"fmt"
)

// ErrAgentBindingValidation marks daemon/runtime binding rejections caused by
// request or registry mismatches, not by storage backend failures.
var ErrAgentBindingValidation = errors.New("riidoaiserver: agent binding validation failed")

type agentBindingValidationError struct {
	message string
}

func (err agentBindingValidationError) Error() string {
	return err.message
}

func (err agentBindingValidationError) Is(target error) bool {
	return target == ErrAgentBindingValidation
}

func newAgentBindingValidationErrorf(format string, args ...any) error {
	return agentBindingValidationError{message: fmt.Sprintf(format, args...)}
}
