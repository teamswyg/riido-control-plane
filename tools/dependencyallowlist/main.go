package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "dependencyallowlist:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("dependencyallowlist", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	contractPath := fs.String("contract", "dependency_allowlist.riido.json", "dependency allowlist contract JSON path")
	moduleDir := fs.String("module", ".", "Go module directory")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	allowlist, err := loadContract(*contractPath)
	if err != nil {
		return err
	}
	modules, err := listModules(*moduleDir)
	if err != nil {
		return err
	}
	report, err := verifyModuleReport(allowlist, modules)
	if err != nil {
		return err
	}
	if *evidenceOut != "" {
		if err := writeEvidence(*evidenceOut, newEvidence(allowlist, report)); err != nil {
			return err
		}
	}
	fmt.Println(report)
	return nil
}
