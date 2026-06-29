package main

import (
	"errors"
	"fmt"
	"strings"
)

func requireFinalContract(final finalContract) error {
	if len(final.ExposedPorts) == 0 {
		return errors.New("final.exposed_ports is required")
	}
	if len(final.Entrypoint) == 0 {
		return errors.New("final.entrypoint is required")
	}
	if len(final.Env) == 0 {
		return errors.New("final.env is required")
	}
	for i, copy := range final.RequiredCopies {
		if err := requireRequiredCopy(i, copy); err != nil {
			return err
		}
	}
	return nil
}

func requireRequiredCopy(i int, requiredCopy requiredCopyContract) error {
	if strings.TrimSpace(requiredCopy.From) == "" {
		return fmt.Errorf("final.required_copies[%d].from is required", i)
	}
	if strings.TrimSpace(requiredCopy.Source) == "" {
		return fmt.Errorf("final.required_copies[%d].source is required", i)
	}
	if strings.TrimSpace(requiredCopy.Destination) == "" {
		return fmt.Errorf("final.required_copies[%d].destination is required", i)
	}
	return nil
}
