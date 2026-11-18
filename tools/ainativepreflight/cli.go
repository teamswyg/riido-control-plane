package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func run(args []string) error {
	fs := flag.NewFlagSet("ainativepreflight", flag.ContinueOnError)
	path := fs.String("task-definition", "", "private ECS task definition JSON")
	name := fs.String("container", "", "configured container name")
	if fs.Parse(args) != nil {
		return errors.New("invalid arguments")
	}
	if *path == "" || *name == "" || fs.NArg() != 0 {
		return errors.New("task-definition and container are required")
	}
	b, err := os.ReadFile(*path)
	if err != nil {
		return errors.New("cannot read private task definition")
	}
	return validate(b, *name)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "BLOCKED:", err)
		os.Exit(1)
	}
	fmt.Println("Production native AI configuration shape: PASS (endpoint ownership and cross-service key equality require private preflight)")
}
