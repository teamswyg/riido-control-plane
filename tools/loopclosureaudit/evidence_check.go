package main

func evidenceChecks(checks []check) []check {
	rows := make([]check, 0, len(checks))
	for _, source := range checks {
		row := source
		row.Contains = append([]string(nil), source.Contains...)
		row.Providers = append([]string(nil), source.Providers...)
		row.Claims = append([]string(nil), source.Claims...)
		rows = append(rows, row)
	}
	return rows
}
