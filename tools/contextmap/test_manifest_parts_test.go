package main

func testOwnedContexts() []ownedContext {
	out := make([]ownedContext, 0, len(requiredOwnedContexts))
	for _, id := range requiredOwnedContexts {
		out = append(out, ownedContext{
			ID:             id,
			Name:           id,
			OwnerPaths:     []string{"owned/" + id},
			Responsibility: "owned",
		})
	}
	return out
}

func testImportedContexts() []importedContext {
	out := make([]importedContext, 0, len(requiredImportedContexts))
	for _, id := range requiredImportedContexts {
		out = append(out, importedContext{
			ID:           id,
			Name:         id,
			ImportedFrom: "fixture",
			Use:          "fixture",
		})
	}
	return out
}

func testExternalContexts() []externalContext {
	out := make([]externalContext, 0, len(requiredExternalContexts))
	for _, id := range requiredExternalContexts {
		out = append(out, externalContext{
			ID:       id,
			Name:     id,
			Owner:    "fixture",
			Boundary: "fixture",
		})
	}
	return out
}
