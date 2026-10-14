package main

import (
	"fmt"
	"strings"
)

type commandInputPaths []string

func (paths *commandInputPaths) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("commands-in path must not be empty")
	}
	*paths = append(*paths, value)
	return nil
}

func (paths commandInputPaths) String() string {
	return strings.Join(paths, ",")
}
