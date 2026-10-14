package main

func hasSurface(items []surface, name string) bool {
	for _, item := range items {
		if item.Name == name && item.Role != "" {
			return true
		}
	}
	return false
}

func hasValue(items []value, want string) bool {
	for _, item := range items {
		if item.Value == want && item.Owner != "" {
			return true
		}
	}
	return false
}

func hasRule(items []rule, id string) bool {
	for _, item := range items {
		if item.ID == id && item.Rule != "" {
			return true
		}
	}
	return false
}
