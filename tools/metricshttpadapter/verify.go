package main

import "fmt"

func verify(root string, m manifest, result adapterResult, checkDoc bool) error {
	if err := verifySourceChecks(root, m.SourceChecks); err != nil {
		return err
	}
	if err := verifyFields(m, result); err != nil {
		return err
	}
	if err := verifyStatuses(m, result); err != nil {
		return err
	}
	if result.HTTPBreakdownRows == 0 || result.StoreBreakdownRows == 0 {
		return fmt.Errorf("missing metric breakdown rows: http=%d store=%d", result.HTTPBreakdownRows, result.StoreBreakdownRows)
	}
	if checkDoc {
		return verifyDoc(root, m)
	}
	return nil
}

func verifyFields(m manifest, result adapterResult) error {
	for _, field := range m.RequiredFields {
		if !result.Fields[field] {
			return fmt.Errorf("missing metrics JSON field %q", field)
		}
	}
	return nil
}
