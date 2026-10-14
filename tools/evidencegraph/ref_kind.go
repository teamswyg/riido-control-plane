package main

var evidenceRefKinds = map[string]struct{}{
	"api":           {},
	"artifact":      {},
	"benchmark":     {},
	"code":          {},
	"config":        {},
	"fixture":       {},
	"generated_doc": {},
	"manifest":      {},
	"projection":    {},
	"script":        {},
	"test":          {},
	"tool":          {},
	"workflow":      {},
}

func knownEvidenceRefKind(kind string) bool {
	_, ok := evidenceRefKinds[kind]
	return ok
}
