// ciworkflowparity verifies that the preserved baseline Go workflow has an
// exact native riido-ci parity candidate without granting retirement authority.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	repoRoot := flag.String("repo-root", ".", "repository root")
	contractPath := flag.String("contract", defaultContract, "baseline parity contract")
	evidenceOut := flag.String("evidence-out", "", "optional mode-0600 evidence output")
	flag.Parse()
	if flag.NArg() != 0 {
		fail(errors.New("unexpected positional arguments"))
	}
	result, err := verify(*repoRoot, *contractPath)
	if err != nil {
		fail(err)
	}
	if *evidenceOut != "" {
		if err := writeEvidence(*evidenceOut, result); err != nil {
			fail(err)
		}
	}
	fmt.Printf("%s decision=%s pipeline=%s adapters=%d retirement_authorized=%t\n",
		result.SchemaVersion, result.Decision, result.PipelineID,
		result.RequiredAdapterCount, result.RetirementAuthorized)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
