package main

import "fmt"

func errUnknownScenario(value string) error {
	return fmt.Errorf("unknown scenario %q", value)
}
