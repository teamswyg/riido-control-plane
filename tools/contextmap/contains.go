package main

func hasOwned(items []ownedContext, id string) bool {
	for _, item := range items {
		if item.ID == id && item.Name != "" && len(item.OwnerPaths) > 0 {
			return true
		}
	}
	return false
}

func hasImported(items []importedContext, id string) bool {
	for _, item := range items {
		if item.ID == id && item.ImportedFrom != "" && item.Use != "" {
			return true
		}
	}
	return false
}

func hasExternal(items []externalContext, id string) bool {
	for _, item := range items {
		if item.ID == id && item.Owner != "" && item.Boundary != "" {
			return true
		}
	}
	return false
}
