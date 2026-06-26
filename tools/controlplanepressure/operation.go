package main

type pressureOperation struct {
	run     func() error
	cleanup func()
}

func newPressureOperation(run func() error) pressureOperation {
	return pressureOperation{run: run}
}

func (op pressureOperation) close() {
	if op.cleanup != nil {
		op.cleanup()
	}
}
