// containercontract validates the riido_ai_server Dockerfile against the
// public executable container image contract.
package main

import (
	"fmt"
	"os"
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
