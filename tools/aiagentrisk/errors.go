package main

import "fmt"

func errUnsafePackagePath(path string) error {
	return fmt.Errorf("unsafe package path %q", path)
}
