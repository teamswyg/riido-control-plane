package main

import "fmt"

func verify(root string, m manifest, shape emfShape, checkDoc bool) error {
	if err := verifySourceChecks(root, m.SourceChecks); err != nil {
		return err
	}
	if err := verifyDimensions(m, shape); err != nil {
		return err
	}
	if err := verifyJSONFields(m, shape); err != nil {
		return err
	}
	if err := verifyMetricUnits(m, shape); err != nil {
		return err
	}
	if checkDoc {
		return verifyDoc(root, m)
	}
	return nil
}

func verifyDimensions(m manifest, shape emfShape) error {
	for _, want := range m.RequiredDimensions {
		if !hasString(shape.Dimensions, want) {
			return fmt.Errorf("missing EMF dimension %q", want)
		}
	}
	return nil
}

func verifyJSONFields(m manifest, shape emfShape) error {
	for _, field := range m.RequiredJSONFields {
		if !shape.JSONFields[field] {
			return fmt.Errorf("missing EMF JSON field %q", field)
		}
	}
	return nil
}

func verifyMetricUnits(m manifest, shape emfShape) error {
	for _, required := range m.RequiredMetricUnit {
		if got := shape.MetricUnits[required.Name]; got != required.Unit {
			return fmt.Errorf("metric %s unit = %q, want %q", required.Name, got, required.Unit)
		}
	}
	return nil
}
