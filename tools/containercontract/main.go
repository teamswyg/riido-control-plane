// containercontract validates the riido_ai_server Dockerfile against the
// public executable container image contract.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	contractSchemaVersion = "riido-container-image-contract.v1"
	checkSchemaVersion    = "riido-container-image-contract-check.v1"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "containercontract:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("containercontract", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	contractPath := fs.String("contract", "", "path to riido-container-image-contract.v1")
	outPath := fs.String("out", "", "optional path to write check JSON, or - for stdout")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*contractPath) == "" {
		return errors.New("-contract is required")
	}
	record, err := loadAndVerify(*contractPath)
	if err != nil {
		return err
	}
	return writeRecord(resolveOutPath(outPath, evidenceOut), record, stdout)
}

func loadAndVerify(contractPath string) (checkRecord, error) {
	contract, err := loadContract(contractPath)
	if err != nil {
		return checkRecord{}, err
	}
	return verifyContract(contract)
}

func resolveOutPath(outPath, evidenceOut *string) string {
	if strings.TrimSpace(*outPath) != "" {
		return *outPath
	}
	return *evidenceOut
}

func writeRecord(path string, record checkRecord, stdout io.Writer) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	body, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if path == "-" {
		_, err := stdout.Write(body)
		return err
	}
	return os.WriteFile(path, body, 0o644)
}
