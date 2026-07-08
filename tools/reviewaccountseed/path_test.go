package main

import "testing"

func TestMustRunReturnsOnSuccessfulCommand(t *testing.T) {
	mustRun([]string{"-repo", "../.."})
}
