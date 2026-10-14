package main

func ownedOwnerFixture(path string) string {
	return `{"generated_doc":"docs/seed.md",` +
		`"workflow":".github/workflows/seed.yml",` +
		`"evidence_artifact":"seed-evidence",` +
		`"seed_ssot":"` + path + `"}`
}

func ownedWorkflowFixture(mode string) string {
	return "" +
		"steps:\n" +
		"  - run: go run ./tools/seedcheck -check-doc -evidence-out out/seed.json\n" +
		"  - uses: actions/upload-artifact@v7\n" +
		"    with:\n" +
		"      name: seed-evidence\n" +
		"      path: out/seed.json\n" +
		"      if-no-files-found: " + mode + "\n"
}
