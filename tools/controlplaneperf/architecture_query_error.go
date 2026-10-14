package main

import "fmt"

func errArchitectureQueryPathRequired() error {
	return fmt.Errorf("architecture query output requires at least one -architecture-query-path")
}
