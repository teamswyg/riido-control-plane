package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if err := runMain(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "aiagentload:", err)
		os.Exit(1)
	}
}

func runMain(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	report, err := run(context.Background(), cfg)
	if err != nil {
		return err
	}
	printSummary(report)
	return writeReport(cfg.OutputPath, report)
}
