package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if err := runMain(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "aiagentthreadsnapshot:", err)
		os.Exit(1)
	}
}

func runMain(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	report, err := capture(context.Background(), cfg)
	if err != nil {
		return err
	}
	return writeReport(cfg.OutputPath, report)
}
