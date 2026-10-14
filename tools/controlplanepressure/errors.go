package main

import "fmt"

type unexpectedStatusError int

func (e unexpectedStatusError) Error() string {
	return fmt.Sprintf("unexpected http status %d", int(e))
}

func errUnexpectedStatus(code int) error {
	return unexpectedStatusError(code)
}
