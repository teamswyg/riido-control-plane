package main

import "fmt"

func required(name string) error {
	return fmt.Errorf("%s is required", name)
}
