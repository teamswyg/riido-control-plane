package riidoaiserver

func validateReadScopePartStatus(parts []*taskReadPartStatus) error {
	seen := map[string]bool{}
	for _, part := range parts {
		if part == nil || part.Name == "" || part.Complete == nil || seen[part.Name] {
			return taskContextReadError{"invalid"}
		}
		seen[part.Name] = true
		switch part.Status {
		case "available":
		case "empty", "not_applicable":
			if !*part.Complete {
				return taskContextReadError{"invalid"}
			}
		case "partial", "unavailable", "unsupported", "deferred":
			if *part.Complete {
				return taskContextReadError{"invalid"}
			}
		default:
			return taskContextReadError{"invalid"}
		}
	}
	return nil
}

func readScopeMetadataComplete(parts []*taskReadPartStatus) bool {
	for _, part := range parts {
		switch part.Name {
		case "document", "description", "document_alternate", "comments", "board":
			continue
		}
		if part.Complete == nil || !*part.Complete {
			return false
		}
	}
	return true
}
