package main

import "fmt"

func unsupportedRefreshCommandSchema(schema string) error {
	return fmt.Errorf("unsupported schema_version %q", schema)
}
