package main

func importedSourceFixture(schema, id string) string {
	return `{"schema_version":"` + schema + `","id":"` + id + `"}`
}

func importedOwnerFixture(schema, id string) string {
	return `{"generated_doc":"docs/projection.md",` +
		`"workflow":".github/workflows/projection.yml",` +
		`"evidence_artifact":"projection-evidence",` +
		`"source_contracts_manifest":{` +
		`"repo":"upstream/contracts",` +
		`"path":"docs/source.riido.json",` +
		`"schema_version":"` + schema + `",` +
		`"id":"` + id + `"}}`
}

func importedWorkflowFixture(mode string) string {
	return "" +
		"steps:\n" +
		"  - run: go run ./tools/projection -check-doc -evidence-out out/projection.json\n" +
		"  - uses: actions/upload-artifact@v4\n" +
		"    with:\n" +
		"      name: projection-evidence\n" +
		"      path: out/projection.json\n" +
		"      if-no-files-found: " + mode + "\n"
}
