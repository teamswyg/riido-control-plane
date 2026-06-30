package main

import "fmt"

func verifyPartialMeasurements(check readinessCheck) error {
	if check.Status != "partial" {
		return nil
	}
	for _, measurement := range check.Measurements {
		if measurement.Kind != "manual" {
			return nil
		}
	}
	return fmt.Errorf("readiness check %s partial evidence must bind a non-manual measurement", check.ID)
}
