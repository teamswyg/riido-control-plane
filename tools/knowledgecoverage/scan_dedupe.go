package main

func dedupeDocs(docs []docClass) []docClass {
	seen := map[string]bool{}
	var out []docClass
	for _, doc := range docs {
		if seen[doc.Path] {
			continue
		}
		seen[doc.Path] = true
		out = append(out, doc)
	}
	return out
}
