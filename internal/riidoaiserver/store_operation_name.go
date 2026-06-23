package riidoaiserver

import "strings"

const unknownStoreOperation = "store_unknown"

type StoreOperationName string

const (
	StoreOperationCreateTask       StoreOperationName = "store_create_task"
	StoreOperationCreateAssignment StoreOperationName = "store_create_assignment"
	StoreOperationCancelAssignment StoreOperationName = "store_cancel_assignment"
	StoreOperationPollAssignment   StoreOperationName = "store_poll_assignment"
	StoreOperationWaitAssignment   StoreOperationName = "store_wait_assignment"
	StoreOperationLeaseAssignment  StoreOperationName = "store_lease_assignment"
	StoreOperationAppendEvent      StoreOperationName = "store_append_event"
)

func (op StoreOperationName) String() string {
	value := strings.TrimSpace(string(op))
	if value == "" {
		return unknownStoreOperation
	}
	return value
}
