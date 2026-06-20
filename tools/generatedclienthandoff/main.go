package main

import (
	"fmt"
	"os"
)

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "generatedclienthandoff:", err)
		os.Exit(1)
	}
}
