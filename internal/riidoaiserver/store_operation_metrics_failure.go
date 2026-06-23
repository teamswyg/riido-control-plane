package riidoaiserver

import (
	"context"
	"errors"
)

func storeOperationFailed(err error) bool {
	return err != nil &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, ErrAgentBindingValidation)
}
