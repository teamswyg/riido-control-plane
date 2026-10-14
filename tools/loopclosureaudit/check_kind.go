package main

type checkKindSpec struct {
	kind    string
	key     func(check) string
	surface func(check, indexes) *proofSurface
	verify  func(string, check, indexes) error
}

func checkKindByName(kind string) (checkKindSpec, bool) {
	for _, spec := range checkKindSpecs {
		if spec.kind == kind {
			return spec, true
		}
	}
	return checkKindSpec{}, false
}
